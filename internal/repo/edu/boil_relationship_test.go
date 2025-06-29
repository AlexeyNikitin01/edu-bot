// Code generated by SQLBoiler 4.18.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package edu

import "testing"

// TestToOne tests cannot be run in parallel
// or deadlocks can occur.
func TestToOne(t *testing.T) {
	t.Run("AnswerToQuestionUsingQuestion", testAnswerToOneQuestionUsingQuestion)
	t.Run("UsersQuestionToQuestionUsingQuestion", testUsersQuestionToOneQuestionUsingQuestion)
	t.Run("UsersQuestionToUserUsingUser", testUsersQuestionToOneUserUsingUser)
}

// TestOneToOne tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOne(t *testing.T) {}

// TestToMany tests cannot be run in parallel
// or deadlocks can occur.
func TestToMany(t *testing.T) {
	t.Run("QuestionToAnswers", testQuestionToManyAnswers)
	t.Run("QuestionToUsersQuestions", testQuestionToManyUsersQuestions)
	t.Run("UserToUsersQuestions", testUserToManyUsersQuestions)
}

// TestToOneSet tests cannot be run in parallel
// or deadlocks can occur.
func TestToOneSet(t *testing.T) {
	t.Run("AnswerToQuestionUsingAnswers", testAnswerToOneSetOpQuestionUsingQuestion)
	t.Run("UsersQuestionToQuestionUsingUsersQuestions", testUsersQuestionToOneSetOpQuestionUsingQuestion)
	t.Run("UsersQuestionToUserUsingUsersQuestions", testUsersQuestionToOneSetOpUserUsingUser)
}

// TestToOneRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestToOneRemove(t *testing.T) {}

// TestOneToOneSet tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOneSet(t *testing.T) {}

// TestOneToOneRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestOneToOneRemove(t *testing.T) {}

// TestToManyAdd tests cannot be run in parallel
// or deadlocks can occur.
func TestToManyAdd(t *testing.T) {
	t.Run("QuestionToAnswers", testQuestionToManyAddOpAnswers)
	t.Run("QuestionToUsersQuestions", testQuestionToManyAddOpUsersQuestions)
	t.Run("UserToUsersQuestions", testUserToManyAddOpUsersQuestions)
}

// TestToManySet tests cannot be run in parallel
// or deadlocks can occur.
func TestToManySet(t *testing.T) {}

// TestToManyRemove tests cannot be run in parallel
// or deadlocks can occur.
func TestToManyRemove(t *testing.T) {}
