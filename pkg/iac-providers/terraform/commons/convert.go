/*
    Copyright (C) 2020 Accurics, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/

package commons

/*
	Following code has been borrowed and modifed from:
	https://github.com/tmccombs/hcl2json/blob/5c1402dc2b410e362afee45a3cf15dcb08bc1f2c/convert.go
*/

import (
	"fmt"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	ctyconvert "github.com/zclconf/go-cty/cty/convert"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type JsonObj map[string]interface{}

type Converter struct {
	Bytes []byte
}

func (c *Converter) rangeSource(r hcl.Range) string {
	return string(c.Bytes[r.Start.Byte:r.End.Byte])
}

func (c *Converter) ConvertBody(body *hclsyntax.Body) (JsonObj, error) {
	var err error
	out := make(JsonObj)
	for key, value := range body.Attributes {
		out[key], err = c.ConvertExpression(value.Expr)
		if err != nil {
			return nil, err
		}
	}

	for _, block := range body.Blocks {
		blockOut := make(JsonObj)
		err = c.convertBlock(block, blockOut)
		if err != nil {
			return nil, err
		}
		blockConfig := blockOut[block.Type].(JsonObj)
		if _, present := out[block.Type]; !present {
			out[block.Type] = []JsonObj{blockConfig}
		} else {
			list := out[block.Type].([]JsonObj)
			list = append(list, blockConfig)
			out[block.Type] = list
		}
	}

	return out, nil
}

func (c *Converter) convertBlock(block *hclsyntax.Block, out JsonObj) error {
	var key string = block.Type

	value, err := c.ConvertBody(block.Body)
	if err != nil {
		return err
	}

	for _, label := range block.Labels {
		if inner, exists := out[key]; exists {
			var ok bool
			out, ok = inner.(JsonObj)
			if !ok {
				// TODO: better diagnostics
				return fmt.Errorf("unable to convert Block to JSON: %v.%v", block.Type, strings.Join(block.Labels, "."))
			}
		} else {
			obj := make(JsonObj)
			out[key] = obj
			out = obj
		}
		key = label
	}

	if current, exists := out[key]; exists {
		if list, ok := current.([]interface{}); ok {
			out[key] = append(list, value)
		} else {
			out[key] = []interface{}{current, value}
		}
	} else {
		out[key] = value
	}

	return nil
}

func (c *Converter) ConvertExpression(expr hclsyntax.Expression) (interface{}, error) {
	// assume it is hcl syntax (because, um, it is)
	switch value := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		return ctyjson.SimpleJSONValue{Value: value.Val}, nil
	case *hclsyntax.TemplateExpr:
		return c.convertTemplate(value)
	case *hclsyntax.TemplateWrapExpr:
		return c.ConvertExpression(value.Wrapped)
	case *hclsyntax.TupleConsExpr:
		var list []interface{}
		for _, ex := range value.Exprs {
			elem, err := c.ConvertExpression(ex)
			if err != nil {
				return nil, err
			}
			list = append(list, elem)
		}
		return list, nil
	case *hclsyntax.ObjectConsExpr:
		m := make(JsonObj)
		for _, item := range value.Items {
			key, err := c.convertKey(item.KeyExpr)
			if err != nil {
				return nil, err
			}
			m[key], err = c.ConvertExpression(item.ValueExpr)
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	default:
		return c.wrapExpr(expr), nil
	}
}

func (c *Converter) convertTemplate(t *hclsyntax.TemplateExpr) (string, error) {
	if t.IsStringLiteral() {
		// safe because the value is just the string
		v, err := t.Value(nil)
		if err != nil {
			return "", err
		}
		return v.AsString(), nil
	}
	var builder strings.Builder
	for _, part := range t.Parts {
		s, err := c.convertStringPart(part)
		if err != nil {
			return "", err
		}
		builder.WriteString(s)
	}
	return builder.String(), nil
}

func (c *Converter) convertStringPart(expr hclsyntax.Expression) (string, error) {
	switch v := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		s, err := ctyconvert.Convert(v.Val, cty.String)
		if err != nil {
			return "", err
		}
		return s.AsString(), nil
	case *hclsyntax.TemplateExpr:
		return c.convertTemplate(v)
	case *hclsyntax.TemplateWrapExpr:
		return c.convertStringPart(v.Wrapped)
	case *hclsyntax.ConditionalExpr:
		return c.convertTemplateConditional(v)
	case *hclsyntax.TemplateJoinExpr:
		return c.convertTemplateFor(v.Tuple.(*hclsyntax.ForExpr))
	default:
		// treating as an embedded expression
		return c.wrapExpr(expr), nil
	}
}

func (c *Converter) convertKey(keyExpr hclsyntax.Expression) (string, error) {
	// a key should never have dynamic input
	if k, isKeyExpr := keyExpr.(*hclsyntax.ObjectConsKeyExpr); isKeyExpr {
		keyExpr = k.Wrapped
		if _, isTraversal := keyExpr.(*hclsyntax.ScopeTraversalExpr); isTraversal {
			return c.rangeSource(keyExpr.Range()), nil
		}
	}
	return c.convertStringPart(keyExpr)
}

func (c *Converter) convertTemplateConditional(expr *hclsyntax.ConditionalExpr) (string, error) {
	var builder strings.Builder
	builder.WriteString("%{if ")
	builder.WriteString(c.rangeSource(expr.Condition.Range()))
	builder.WriteString("}")
	trueResult, err := c.convertStringPart(expr.TrueResult)
	if err != nil {
		return "", nil
	}
	builder.WriteString(trueResult)
	falseResult, _ := c.convertStringPart(expr.FalseResult)
	if len(falseResult) > 0 {
		builder.WriteString("%{else}")
		builder.WriteString(falseResult)
	}
	builder.WriteString("%{endif}")

	return builder.String(), nil
}

func (c *Converter) convertTemplateFor(expr *hclsyntax.ForExpr) (string, error) {
	var builder strings.Builder
	builder.WriteString("%{for ")
	if len(expr.KeyVar) > 0 {
		builder.WriteString(expr.KeyVar)
		builder.WriteString(", ")
	}
	builder.WriteString(expr.ValVar)
	builder.WriteString(" in ")
	builder.WriteString(c.rangeSource(expr.CollExpr.Range()))
	builder.WriteString("}")
	templ, err := c.convertStringPart(expr.ValExpr)
	if err != nil {
		return "", err
	}
	builder.WriteString(templ)
	builder.WriteString("%{endfor}")

	return builder.String(), nil
}

func (c *Converter) wrapExpr(expr hclsyntax.Expression) string {
	return "${" + c.rangeSource(expr.Range()) + "}"
}
