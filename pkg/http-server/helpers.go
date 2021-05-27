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

package httpserver

import (
	"fmt"
	"github.com/accurics/terrascan/pkg/writer"
	"go.uber.org/zap"
	"net/http"
)

// apiResponse creates an API response
func apiResponse(w http.ResponseWriter, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, msg)
}

// apiErrorResponse creates an API error response
func apiErrorResponse(w http.ResponseWriter, errMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.Error(w, errMsg, statusCode)
}

func apiSarifResponse(w http.ResponseWriter, output interface{}, statusCode int ) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	writer.SarifWriter(output, w)
}

func handleSarifValidation(w http.ResponseWriter) {
	errMsg := "config_only feature is not allowed with sarif output format"
	zap.S().Error(errMsg)
	apiErrorResponse(w, errMsg, http.StatusBadRequest)
}
