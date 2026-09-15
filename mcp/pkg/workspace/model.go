// Package workspace は workspace.yaml（Definition）と workspace.local.yaml（Overlay）の
// モデル・検証・所在解決を提供する。MCP サーバと validate サブコマンドが共用する。
package workspace

// Workspace は workspace.yaml の最上位構造（Definition）。
// 場所(clone先)や再現性は持たない。再現性は各リポの git tag / go.sum が担う。
type Workspace struct {
	SchemaVersion int            `yaml:"schemaVersion"`
	Workspace     string         `yaml:"workspace"`
	Description   string         `yaml:"description"`
	Roles         []Role         `yaml:"roles"`
	Repos         []Repo         `yaml:"repos"`
	Relationships []Relationship `yaml:"relationships"`
	Skills        []SkillBinding `yaml:"skills"`
}

// Role は作成時に宣言する自由語彙。固定 enum ではない。
type Role struct {
	ID          string `yaml:"id" json:"id"`
	Description string `yaml:"description" json:"description,omitempty"`
}

// Repo はワークスペースを構成するリポジトリ。
// revision は追跡ブランチの指定（既定はデフォルトブランチ、無指定も正常）。
type Repo struct {
	Name     string `yaml:"name"`
	Repo     string `yaml:"repo"`
	Role     string `yaml:"role"`
	Revision string `yaml:"revision"`
	Optional bool   `yaml:"optional"`
}

// Relationship はリポ間の役割関係。from/to は Repo.Name を参照する。
type Relationship struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
	Type string `yaml:"type"`
}

// SkillBinding は「何をするとき、どのスキルか」の束縛。
type SkillBinding struct {
	When  string   `yaml:"when" json:"when,omitempty"`
	Use   string   `yaml:"use" json:"use"`
	Roles []string `yaml:"roles" json:"roles,omitempty"`
}

// Overlay は workspace.local.yaml（非コミットの一時上書き）。go.work の replace に相当。
type Overlay struct {
	Overrides []Override `yaml:"overrides"`
}

// Override は特定リポの revision をローカルで差し替える。worktree に展開して参照する。
type Override struct {
	Name     string `yaml:"name"`
	Revision string `yaml:"revision"`
}
