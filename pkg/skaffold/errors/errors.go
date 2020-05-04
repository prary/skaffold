/*
Copyright 2020 The Skaffold Authors

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

package errors

import (
	"github.com/GoogleContainerTools/skaffold/proto"
)

const (
	Build       = phase("Build")
	Deploy      = phase("Deploy")
	StatusCheck = phase("StatusCheck")
	FileSync    = phase("FileSync")
)

type phase string

func ErrorCodeFromError(_ error, _ phase) proto.ErrorCode {
	return proto.ErrorCode_COULD_NOT_DETERMINE
}
