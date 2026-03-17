# parry-bossrush

Go + Ebitengine で作った、**パリィ重視の固定画面2Dボス戦アクション**のMVPです。

- 画像/音声アセットなし
- プレイヤーは左右移動 + パリィのみ
- パリィ成功で自動反撃してボスにダメージ
- 矩形と色だけで状態を可視化

## 必要環境

- Go 1.22 以上
- OS の GUI 環境（Ebitengine のウィンドウを開くため）

## 依存取得

```bash
go mod tidy
```

## 開発用の起動方法

```bash
go run .
```

## ビルド方法

```bash
go build -o boss-parry .
```

## ビルド後の実行方法

### macOS / Linux

```bash
./boss-parry
```

### Windows (PowerShell)

```powershell
.\boss-parry.exe
```

## 操作方法

- `Left / Right`: 左右移動
- `Space`: パリィ
- `R`: リトライ（Battle中、勝敗表示後）
- `Esc`: 画面戻る（Battle -> Boss Select、Boss Select -> Title）

## 画面構成

1. **Title**
   - `Press Space to Start` で Boss Select へ
2. **Boss Select**
   - 現状は1体のみ（Training Sentinel）
   - `Space` で Battle 開始
   - `Esc` で Title に戻る
3. **Battle**
   - プレイヤー、ボス、HP、攻撃判定を表示
   - 勝利時 `VICTORY`、敗北時 `GAME OVER`
   - `R` で再戦

## MVPルール

- プレイヤーHPは3、ボスHPは5
- ボスは距離が遠いと接近、近いと攻撃準備へ
- 攻撃は1種類のみ（予兆 -> 攻撃 -> 硬直）
- 攻撃中は攻撃判定ボックスを表示
- 予兆後の攻撃ヒットタイミングに `Space` を合わせるとパリィ成功
- パリィ成功時:
  - 自動反撃でボスHPが1減る
  - 短いヒットストップが入る
  - ボスがスタンする
- パリィ失敗/通常被弾時:
  - プレイヤーHPが1減る
  - 短時間無敵

## 状態色

- プレイヤー
  - 通常: 白
  - パリィ受付中: 青
  - 被弾無敵中: 赤
- ボス
  - 通常: 灰
  - 攻撃予兆中: 黄
  - 攻撃中: 赤
  - スタン中: 紫
- 攻撃判定
  - 攻撃中のみオレンジ矩形で表示
