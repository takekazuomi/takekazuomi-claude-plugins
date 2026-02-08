# スキルテンプレート管理方法の調査結果

## 概要

Claude Codeスキル作成時のコードテンプレート管理方法（Markdownコードブロック vs 別ファイル）について調査。

## 調査結果

### 公式推奨（code.claude.com/docs/en/skills.md）

| 方法 | 利用場面 | トークン効率 | メンテナンス性 |
| ------ | --------- | ------------ | -------------- |
| **Markdownコードブロック** | <200行のテンプレート | 中程度 | 高い（一箇所管理） |
| **別ファイル（参照）** | 詳細ドキュメント、リファレンス | 高い（需要時読込） | 高い |
| **スクリプト bundled** | テンプレート生成、複雑な変換処理 | 最高（実行のみ） | 良い |

### 公式ガイドライン

1. **SKILL.mdは500行以下**を推奨（パフォーマンス最適化）
2. **Progressive Disclosure パターン**：
   - SKILL.md に本質的な情報を集約
   - 詳細はリンク経由で参照ファイルに分割
   - スクリプトは実行するだけ（内容は読み込まない）

### テンプレートサイズ別の推奨

#### 小規模（<200行）: Markdownコードブロック

- コンテキスト内に完全に含まれる
- Claudeがすぐに実行できる
- 環境構築が不要

#### 大規模（>200行）: 別ファイル化

- `scripts/` ディレクトリにスクリプト配置
- SKILL.mdから実行コマンドのみ記載
- トークン効率が向上

### 公式の複雑スキル例（codebase-visualizer）

```text
my-skill/
├── SKILL.md (overview - 500行以下)
├── reference.md (詳細ドキュメント)
├── examples.md (使用例)
└── scripts/
    └── helper.py (実行スクリプト)
```

SKILL.mdでの参照方法：

````markdown
Run the visualization script:

```bash
python ~/.claude/skills/codebase-visualizer/scripts/visualize.py .
```
````

## 現状のプロジェクト評価

### go-new スキル（~207行）

- Makefileテンプレート：~52行
- README.mdテンプレート：~27行
- .gitignoreテンプレート：~11行
- .golangci.ymlテンプレート：~60行

**評価: 適切** - 各テンプレートが200行未満

### mysql-local スキル（~152行）

- docker-compose.yml：~27行
- 初期化SQL：~12行
- .envrc.sample：~7行
- Makefileターゲット：~16行

**評価: 適切** - シンプルな構造

## 結論

現在のアプローチ（Markdownコードブロック）は公式推奨に沿っている。

- 各テンプレートが200行未満のため、コードブロックが適切
- SKILL.md全体も500行以下でパフォーマンス基準を満たす
- メンテナンス性が高い（一箇所管理）

## 将来的な拡張時の推奨

テンプレートが大規模化した場合：

```text
mysql-local/
├── SKILL.md (概要と基本手順)
├── examples/
│   ├── schema.sql (テーブル定義例)
│   └── seed-data.sql (サンプルデータ)
└── scripts/
    └── backup.sh (バックアップユーティリティ)
```

## 参考ドキュメント

- Skillsガイド: <https://code.claude.com/docs/en/skills.md>
- Progressive Disclosure: 公式ドキュメントの「Add supporting files」セクション
