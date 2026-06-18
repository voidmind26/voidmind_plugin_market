package db

import "testing"

func TestIsAllowedWriteQuery(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want bool
	}{
		{name: "allow insert", sql: "INSERT INTO t_user(id) VALUES (1)", want: true},
		{name: "allow update", sql: "UPDATE t_user SET name = 'a' WHERE id = 1", want: true},
		{name: "allow create", sql: "CREATE TABLE t_user(id bigint primary key)", want: true},
		{name: "allow alter", sql: "ALTER TABLE t_user ADD COLUMN age int", want: true},
		{name: "reject delete", sql: "DELETE FROM t_user WHERE id = 1", want: false},
		{name: "reject drop", sql: "DROP TABLE t_user", want: false},
		{name: "reject truncate", sql: "TRUNCATE TABLE t_user", want: false},
		{name: "reject replace", sql: "REPLACE INTO t_user(id) VALUES (1)", want: false},
		{name: "reject multiple statements", sql: "INSERT INTO t_user(id) VALUES (1); DELETE FROM t_user WHERE id = 1", want: false},
		{name: "reject blank", sql: "   ", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllowedWriteQuery(tt.sql); got != tt.want {
				t.Fatalf("isAllowedWriteQuery(%q) = %v, want %v", tt.sql, got, tt.want)
			}
		})
	}
}

func TestIsReadOnlyQuery(t *testing.T) {
	if !isReadOnlyQuery("SELECT * FROM t_user") {
		t.Fatal("expected select to be read only")
	}
	if isReadOnlyQuery("UPDATE t_user SET name='a'") {
		t.Fatal("did not expect update to be read only")
	}
}

func TestDetectSQLAction(t *testing.T) {
	action, ok := detectSQLAction("ALTER TABLE t_user ADD COLUMN age int;")
	if !ok {
		t.Fatal("expected action detection to succeed")
	}
	if action != "alter" {
		t.Fatalf("detectSQLAction returned %q, want alter", action)
	}

	if _, ok := detectSQLAction("INSERT INTO t_user(id) VALUES (1); DELETE FROM t_user"); ok {
		t.Fatal("expected multi statement SQL to be rejected")
	}
}

func TestExecuteWriteQueryRejectsBlockedSQL(t *testing.T) {
	tool := &MySQLQueryTool{
		name: "test_db",
	}

	result, err := tool.ExecuteWriteQuery(t.Context(), "DELETE FROM t_user WHERE id = 1", nil)
	if err != nil {
		t.Fatalf("ExecuteWriteQuery returned unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Success {
		t.Fatal("expected blocked write SQL to be rejected")
	}
	if result.Message == "" {
		t.Fatal("expected rejection message")
	}
}
