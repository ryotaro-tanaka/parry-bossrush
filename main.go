package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenW = 640
	screenH = 360
	groundY = 290
)

type ScreenState int

const (
	ScreenTitle ScreenState = iota
	ScreenBossSelect
	ScreenBattle
)

type PlayerState int

const (
	PlayerNormal PlayerState = iota
	PlayerParry
	PlayerInvincible
)

type BossState int

const (
	BossIdle BossState = iota
	BossApproach
	BossTelegraph
	BossAttack
	BossRecovery
	BossStunned
)

type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

func (r Rect) Intersects(other Rect) bool {
	return r.X < other.X+other.W && r.X+r.W > other.X && r.Y < other.Y+other.H && r.Y+r.H > other.Y
}

type Player struct {
	X               float64
	Y               float64
	W               float64
	H               float64
	Speed           float64
	State           PlayerState
	ParryTimer      int
	InvincibleTimer int
	HP              int
}

func (p *Player) Rect() Rect {
	return Rect{X: p.X, Y: p.Y, W: p.W, H: p.H}
}

type Boss struct {
	X          float64
	Y          float64
	W          float64
	H          float64
	Speed      float64
	State      BossState
	StateTimer int
	Facing     float64
	HP         int
}

func (b *Boss) Rect() Rect {
	return Rect{X: b.X, Y: b.Y, W: b.W, H: b.H}
}

func (b *Boss) AttackRect() Rect {
	attackW := 60.0
	offset := 12.0
	if b.Facing >= 0 {
		return Rect{X: b.X + b.W + offset, Y: b.Y + 8, W: attackW, H: b.H - 16}
	}
	return Rect{X: b.X - attackW - offset, Y: b.Y + 8, W: attackW, H: b.H - 16}
}

type Game struct {
	screenState   ScreenState
	player        Player
	boss          Boss
	hitstopTimer  int
	parryResolved bool
	resultText    string
}

func NewGame() *Game {
	g := &Game{screenState: ScreenTitle}
	g.resetBattle()
	return g
}

func (g *Game) resetBattle() {
	g.player = Player{X: 120, Y: groundY - 52, W: 28, H: 52, Speed: 2.4, State: PlayerNormal, HP: 3}
	g.boss = Boss{X: 470, Y: groundY - 70, W: 42, H: 70, Speed: 1.4, State: BossApproach, Facing: -1, HP: 5}
	g.hitstopTimer = 0
	g.parryResolved = false
	g.resultText = ""
}

func (g *Game) Update() error {
	switch g.screenState {
	case ScreenTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.screenState = ScreenBossSelect
		}
	case ScreenBossSelect:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.screenState = ScreenTitle
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.resetBattle()
			g.screenState = ScreenBattle
		}
	case ScreenBattle:
		g.updateBattle()
	}
	return nil
}

func (g *Game) updateBattle() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.screenState = ScreenBossSelect
		return
	}

	if g.resultText != "" {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.resetBattle()
		}
		return
	}

	if g.hitstopTimer > 0 {
		g.hitstopTimer--
		return
	}

	g.updatePlayer()
	g.updateBoss()
	g.resolveAttackHit()

	if g.player.HP <= 0 {
		g.resultText = "GAME OVER"
	}
	if g.boss.HP <= 0 {
		g.resultText = "VICTORY"
	}
}

func (g *Game) updatePlayer() {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.X -= g.player.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.X += g.player.Speed
	}
	g.player.X = math.Max(0, math.Min(g.player.X, screenW-g.player.W))

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.player.State = PlayerParry
		g.player.ParryTimer = 10
	}

	if g.player.ParryTimer > 0 {
		g.player.ParryTimer--
		if g.player.ParryTimer == 0 && g.player.State == PlayerParry {
			g.player.State = PlayerNormal
		}
	}

	if g.player.InvincibleTimer > 0 {
		g.player.InvincibleTimer--
		if g.player.InvincibleTimer == 0 && g.player.State == PlayerInvincible {
			g.player.State = PlayerNormal
		}
	}
}

func (g *Game) updateBoss() {
	dx := (g.player.X + g.player.W/2) - (g.boss.X + g.boss.W/2)
	dist := math.Abs(dx)
	if dx >= 0 {
		g.boss.Facing = 1
	} else {
		g.boss.Facing = -1
	}

	switch g.boss.State {
	case BossIdle:
		g.boss.State = BossApproach
	case BossApproach:
		if dist > 120 {
			if dx > 0 {
				g.boss.X += g.boss.Speed
			} else {
				g.boss.X -= g.boss.Speed
			}
		} else {
			g.boss.State = BossTelegraph
			g.boss.StateTimer = 36
		}
	case BossTelegraph:
		g.boss.StateTimer--
		if g.boss.StateTimer <= 0 {
			g.boss.State = BossAttack
			g.boss.StateTimer = 18
			g.parryResolved = false
		}
	case BossAttack:
		g.boss.StateTimer--
		if g.boss.StateTimer <= 0 {
			g.boss.State = BossRecovery
			g.boss.StateTimer = 24
		}
	case BossRecovery:
		g.boss.StateTimer--
		if g.boss.StateTimer <= 0 {
			g.boss.State = BossApproach
		}
	case BossStunned:
		g.boss.StateTimer--
		if g.boss.StateTimer <= 0 {
			g.boss.State = BossRecovery
			g.boss.StateTimer = 20
		}
	}

	g.boss.X = math.Max(0, math.Min(g.boss.X, screenW-g.boss.W))
}

func (g *Game) resolveAttackHit() {
	if g.boss.State != BossAttack || g.parryResolved {
		return
	}

	if !g.player.Rect().Intersects(g.boss.AttackRect()) {
		return
	}

	if g.player.State == PlayerParry {
		g.parryResolved = true
		g.player.State = PlayerNormal
		g.player.ParryTimer = 0
		g.boss.State = BossStunned
		g.boss.StateTimer = 30
		g.boss.HP--
		g.hitstopTimer = 5
		return
	}

	if g.player.InvincibleTimer == 0 {
		g.player.HP--
		g.player.State = PlayerInvincible
		g.player.InvincibleTimer = 35
		g.parryResolved = true
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{25, 25, 30, 255})
	switch g.screenState {
	case ScreenTitle:
		g.drawTitle(screen)
	case ScreenBossSelect:
		g.drawBossSelect(screen)
	case ScreenBattle:
		g.drawBattle(screen)
	}
}

func (g *Game) drawTitle(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "PARRY BOSSRUSH MVP", 220, 130)
	ebitenutil.DebugPrintAt(screen, "Press Space to Start", 245, 170)
	ebitenutil.DebugPrintAt(screen, "Esc: Quit", 285, 200)
}

func (g *Game) drawBossSelect(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "BOSS SELECT", 275, 80)
	ebitenutil.DebugPrintAt(screen, "> Training Sentinel", 235, 145)
	ebitenutil.DebugPrintAt(screen, "Space: Start Battle", 245, 200)
	ebitenutil.DebugPrintAt(screen, "Esc: Back", 280, 225)
}

func (g *Game) drawBattle(screen *ebiten.Image) {
	ground := ebiten.NewImage(screenW, screenH-groundY)
	ground.Fill(color.RGBA{45, 45, 50, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, groundY)
	screen.DrawImage(ground, op)

	pCol := color.RGBA{240, 240, 240, 255}
	switch g.player.State {
	case PlayerParry:
		pCol = color.RGBA{70, 130, 255, 255}
	case PlayerInvincible:
		pCol = color.RGBA{240, 70, 70, 255}
	}
	drawRect(screen, g.player.Rect(), pCol)

	bCol := color.RGBA{145, 145, 145, 255}
	switch g.boss.State {
	case BossTelegraph:
		bCol = color.RGBA{235, 215, 70, 255}
	case BossAttack:
		bCol = color.RGBA{225, 60, 60, 255}
	case BossStunned:
		bCol = color.RGBA{120, 80, 225, 255}
	}
	drawRect(screen, g.boss.Rect(), bCol)

	if g.boss.State == BossAttack {
		drawRect(screen, g.boss.AttackRect(), color.RGBA{255, 120, 50, 220})
	}

	ebitenutil.DebugPrintAt(screen, "Player HP: "+itoa(g.player.HP), 20, 12)
	ebitenutil.DebugPrintAt(screen, "Boss HP: "+itoa(g.boss.HP), 530, 12)
	ebitenutil.DebugPrintAt(screen, "Move: Left/Right  Parry: Space", 20, 330)
	ebitenutil.DebugPrintAt(screen, "R: Retry  Esc: Back", 460, 330)

	if g.hitstopTimer > 0 {
		ebitenutil.DebugPrintAt(screen, "PARRY!", 300, 120)
	}
	if g.resultText != "" {
		ebitenutil.DebugPrintAt(screen, g.resultText, 285, 150)
		ebitenutil.DebugPrintAt(screen, "Press R to Retry", 260, 180)
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenW, screenH
}

func drawRect(screen *ebiten.Image, r Rect, col color.Color) {
	img := ebiten.NewImage(int(r.W), int(r.H))
	img.Fill(col)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(r.X, r.Y)
	screen.DrawImage(img, op)
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	buf := [12]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func main() {
	ebiten.SetWindowSize(screenW*2, screenH*2)
	ebiten.SetWindowTitle("Parry Bossrush MVP")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
