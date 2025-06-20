// Code generated by SQLBoiler 4.18.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package edu

import "testing"

// This test suite runs each operation test in parallel.
// Example, if your database has 3 tables, the suite will run:
// table1, table2 and table3 Delete in parallel
// table1, table2 and table3 Insert in parallel, and so forth.
// It does NOT run each operation group in parallel.
// Separating the tests thusly grants avoidance of Postgres deadlocks.
func TestParent(t *testing.T) {
	t.Run("Answers", testAnswers)
	t.Run("Questions", testQuestions)
	t.Run("Users", testUsers)
	t.Run("UsersQuestions", testUsersQuestions)
}

func TestDelete(t *testing.T) {
	t.Run("Answers", testAnswersDelete)
	t.Run("Questions", testQuestionsDelete)
	t.Run("Users", testUsersDelete)
	t.Run("UsersQuestions", testUsersQuestionsDelete)
}

func TestQueryDeleteAll(t *testing.T) {
	t.Run("Answers", testAnswersQueryDeleteAll)
	t.Run("Questions", testQuestionsQueryDeleteAll)
	t.Run("Users", testUsersQueryDeleteAll)
	t.Run("UsersQuestions", testUsersQuestionsQueryDeleteAll)
}

func TestSliceDeleteAll(t *testing.T) {
	t.Run("Answers", testAnswersSliceDeleteAll)
	t.Run("Questions", testQuestionsSliceDeleteAll)
	t.Run("Users", testUsersSliceDeleteAll)
	t.Run("UsersQuestions", testUsersQuestionsSliceDeleteAll)
}

func TestExists(t *testing.T) {
	t.Run("Answers", testAnswersExists)
	t.Run("Questions", testQuestionsExists)
	t.Run("Users", testUsersExists)
	t.Run("UsersQuestions", testUsersQuestionsExists)
}

func TestFind(t *testing.T) {
	t.Run("Answers", testAnswersFind)
	t.Run("Questions", testQuestionsFind)
	t.Run("Users", testUsersFind)
	t.Run("UsersQuestions", testUsersQuestionsFind)
}

func TestBind(t *testing.T) {
	t.Run("Answers", testAnswersBind)
	t.Run("Questions", testQuestionsBind)
	t.Run("Users", testUsersBind)
	t.Run("UsersQuestions", testUsersQuestionsBind)
}

func TestOne(t *testing.T) {
	t.Run("Answers", testAnswersOne)
	t.Run("Questions", testQuestionsOne)
	t.Run("Users", testUsersOne)
	t.Run("UsersQuestions", testUsersQuestionsOne)
}

func TestAll(t *testing.T) {
	t.Run("Answers", testAnswersAll)
	t.Run("Questions", testQuestionsAll)
	t.Run("Users", testUsersAll)
	t.Run("UsersQuestions", testUsersQuestionsAll)
}

func TestCount(t *testing.T) {
	t.Run("Answers", testAnswersCount)
	t.Run("Questions", testQuestionsCount)
	t.Run("Users", testUsersCount)
	t.Run("UsersQuestions", testUsersQuestionsCount)
}

func TestHooks(t *testing.T) {
	t.Run("Answers", testAnswersHooks)
	t.Run("Questions", testQuestionsHooks)
	t.Run("Users", testUsersHooks)
	t.Run("UsersQuestions", testUsersQuestionsHooks)
}

func TestInsert(t *testing.T) {
	t.Run("Answers", testAnswersInsert)
	t.Run("Answers", testAnswersInsertWhitelist)
	t.Run("Questions", testQuestionsInsert)
	t.Run("Questions", testQuestionsInsertWhitelist)
	t.Run("Users", testUsersInsert)
	t.Run("Users", testUsersInsertWhitelist)
	t.Run("UsersQuestions", testUsersQuestionsInsert)
	t.Run("UsersQuestions", testUsersQuestionsInsertWhitelist)
}

func TestReload(t *testing.T) {
	t.Run("Answers", testAnswersReload)
	t.Run("Questions", testQuestionsReload)
	t.Run("Users", testUsersReload)
	t.Run("UsersQuestions", testUsersQuestionsReload)
}

func TestReloadAll(t *testing.T) {
	t.Run("Answers", testAnswersReloadAll)
	t.Run("Questions", testQuestionsReloadAll)
	t.Run("Users", testUsersReloadAll)
	t.Run("UsersQuestions", testUsersQuestionsReloadAll)
}

func TestSelect(t *testing.T) {
	t.Run("Answers", testAnswersSelect)
	t.Run("Questions", testQuestionsSelect)
	t.Run("Users", testUsersSelect)
	t.Run("UsersQuestions", testUsersQuestionsSelect)
}

func TestUpdate(t *testing.T) {
	t.Run("Answers", testAnswersUpdate)
	t.Run("Questions", testQuestionsUpdate)
	t.Run("Users", testUsersUpdate)
	t.Run("UsersQuestions", testUsersQuestionsUpdate)
}

func TestSliceUpdateAll(t *testing.T) {
	t.Run("Answers", testAnswersSliceUpdateAll)
	t.Run("Questions", testQuestionsSliceUpdateAll)
	t.Run("Users", testUsersSliceUpdateAll)
	t.Run("UsersQuestions", testUsersQuestionsSliceUpdateAll)
}
