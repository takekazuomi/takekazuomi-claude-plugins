---
description: 社内提案書の依頼で formal（常体）が選ばれるか。構成案の中身の妥当性は測らない
tags: [output, routing]
max_turns: 30
timeout_seconds: 600
allowed_tools: [Read, Glob, Grep, Skill, Write, Bash]
---

社内向けに、APIゲートウェイの認証方式を切り替える提案書を書きたい。
レビューするのは基盤チームのリードです。主張は「API キー認証をやめて OIDC のトークン検証に切り替える」です。
確認は不要なので、構成案を proposal.md に書いてください。
