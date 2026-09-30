package service

import (
	"api-students/app/model"
	"strings"
)

func ApplyStudentPatch(current model.Student, req model.PatchStudentRequest) model.Student {
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}
	if req.Grade != nil {
		current.Grade = *req.Grade
	}
	return current
}

func IsEmptyStudentPatch(req model.PatchStudentRequest) bool {
	return req.Name == nil && req.Grade == nil
}