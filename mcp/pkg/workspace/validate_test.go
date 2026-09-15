package workspace

import "testing"

func validWorkspace() *Workspace {
	return &Workspace{
		SchemaVersion: 1,
		Workspace:     "payments",
		Roles:         []Role{{ID: "source"}, {ID: "api-definition"}},
		Repos: []Repo{
			{Name: "svc", Repo: "github.com/acme/svc", Role: "source"},
			{Name: "proto", Repo: "github.com/acme/proto", Role: "api-definition"},
		},
		Relationships: []Relationship{{From: "svc", To: "proto", Type: "implements"}},
	}
}

func TestValidateOK(t *testing.T) {
	errs, warns := Validate(validWorkspace())
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if len(warns) != 0 {
		t.Fatalf("expected no warnings, got %v", warns)
	}
}

func TestValidateUndeclaredRole(t *testing.T) {
	ws := validWorkspace()
	ws.Repos[0].Role = "ghost"
	errs, _ := Validate(ws)
	if len(errs) == 0 {
		t.Fatal("expected error for undeclared role")
	}
}

func TestValidateDuplicateName(t *testing.T) {
	ws := validWorkspace()
	ws.Repos[1].Name = "svc"
	errs, _ := Validate(ws)
	if len(errs) == 0 {
		t.Fatal("expected error for duplicated repo name")
	}
}

func TestValidateUnknownRelationshipEndpoint(t *testing.T) {
	ws := validWorkspace()
	ws.Relationships[0].To = "missing"
	errs, _ := Validate(ws)
	if len(errs) == 0 {
		t.Fatal("expected error for unknown relationship endpoint")
	}
}

func TestValidateUnusedRoleWarns(t *testing.T) {
	ws := validWorkspace()
	ws.Roles = append(ws.Roles, Role{ID: "documentation"})
	errs, warns := Validate(ws)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	if len(warns) == 0 {
		t.Fatal("expected warning for unused role")
	}
}

// revision は追跡ブランチの指定。無指定もブランチ名も警告しない（再現性は責務外）。
func TestValidateRevisionNeverWarns(t *testing.T) {
	ws := validWorkspace()
	ws.Repos[0].Revision = ""            // 無指定
	ws.Repos[1].Revision = "feature/api" // ブランチ名
	_, warns := Validate(ws)
	if len(warns) != 0 {
		t.Fatalf("revision should never warn, got %v", warns)
	}
}
