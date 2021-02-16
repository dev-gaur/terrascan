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

package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/accurics/terrascan/pkg/config"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	gitConfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var (
	basePath       = config.GetPolicyRepoPath()
	basePolicyPath = config.GetPolicyBasePath()
	repoURL        = config.GetPolicyRepoURL()
	branch         = config.GetPolicyBranch()
)

// Run initializes terrascan if not done already
func Run(isScanCmd bool) error {

	zap.S().Debug("initializing terrascan")

	// check if policy paths exist
	if path, err := os.Stat(basePolicyPath); err == nil && path.IsDir() {
		if isScanCmd {
			return nil
		}
	}

	// download policies
	if err := DownloadPolicies(); err != nil {
		return err
	}

	zap.S().Debug("intialized successfully")
	return nil
}

// DownloadPolicies clones the policies to a local folder
func DownloadPolicies() error {

	tempPath := filepath.Join(os.TempDir(), "terrascan")

	// ensure a clear tempPath
	if err := os.RemoveAll(tempPath); err != nil {
		zap.S().Errorf("failed to clean the existing temporary directory at '%s'. error: '%v'", tempPath, err)
		return err
	}

	// clone the repo
	r, err := git.PlainClone(tempPath, false, &git.CloneOptions{
		URL: repoURL,
	})
	if err != nil {
		zap.S().Errorf("failed to download policies. error: '%v'", err)
		return err
	}

	// create working tree
	w, err := r.Worktree()
	if err != nil {
		zap.S().Errorf("failed to create working tree. error: '%v'", err)
		return err
	}

	// fetch references
	err = r.Fetch(&git.FetchOptions{
		RefSpecs: []gitConfig.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
	})
	if err != nil {
		zap.S().Errorf("failed to fetch references from repo. error: '%v'", err)
		return err
	}

	// checkout policies branch
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		Force:  true,
	})
	if err != nil {
		zap.S().Errorf("failed to checkout branch '%v'. error: '%v'", branch, err)
		return err
	}

	// cleaning the existing cached policies at basePath
	if err = os.RemoveAll(basePath); err != nil {
		zap.S().Errorf("failed to clean the existing policy repository at '%s'. error: '%v'", basePath, err)
		return err
	}

	// move the freshly cloned repo from tempPath to basePath
	if err = os.Rename(tempPath, basePath); err != nil {
		zap.S().Errorf("failed to move the freshly cloned repository from '%s' to '%s'. error: '%v'", tempPath, basePath, err)
		return err
	}

	return nil
}
