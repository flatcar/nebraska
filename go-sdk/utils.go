package nebraska

import (
	"net/http"

	"github.com/kinvolk/nebraska/backend/pkg/codegen"
)

func convertReqEditors(reqEditors ...RequestEditorFn) []codegen.RequestEditorFn {
	var codegenReqEditors []codegen.RequestEditorFn
	for _, reqEditor := range reqEditors {
		codegenReqEditors = append(codegenReqEditors, codegen.RequestEditorFn(reqEditor))
	}
	return codegenReqEditors
}

type rawResponse struct {
	resp *http.Response
}

type commonOptions struct {
	RequestEditors []RequestEditorFn
}

func (r *rawResponse) Response() *http.Response {
	return r.resp
}
