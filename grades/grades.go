package grades

import (
	"fmt"
	"sync"
)

type GradeType string

const (
	GradeTest     = GradeType("Test")
	GradeHomework = GradeType("Homework")
	GradeQuiz     = GradeType("Quiz")
)

type Student struct {
	ID                  int
	FirstName, LastName string
	Grades              []Grade
}

func (s Student) Average() float32 {
	var result float32
	for _, grade := range s.Grades {
		result += grade.Score
	}
	return result / float32(len(s.Grades))
}

type Grade struct {
	Title string
	Type  GradeType
	Score float32
}

type Students []Student

func (s Students) GetByID(id int) (*Student, error) {
	for i := range s {
		if s[i].ID == id {
			return &s[i], nil
		}
	}

	return nil, fmt.Errorf("student with ID %v not found", id)
}

var (
	studentsMutex sync.Mutex
	students      Students
)
