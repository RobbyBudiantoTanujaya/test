package chess

type GameStatus int

const (
	Playing GameStatus = iota
	InCheck
	Checkmate
	Stalemate
	Draw50Move
	DrawInsufficient
)

func (s GameStatus) String() string {
	switch s {
	case Playing:
		return "playing"
	case InCheck:
		return "check"
	case Checkmate:
		return "checkmate"
	case Stalemate:
		return "stalemate"
	case Draw50Move:
		return "draw50"
	case DrawInsufficient:
		return "drawInsufficient"
	}
	return "playing"
}

type Game struct {
	History []*Position
	Moves   []Move
	Status  GameStatus
}

func NewGame() *Game {
	g := &Game{
		History: []*Position{NewStartPosition()},
	}
	g.updateStatus()
	return g
}

func (g *Game) Current() *Position {
	return g.History[len(g.History)-1]
}

func (g *Game) LegalMoves() []Move {
	return g.Current().GenerateLegalMoves()
}

func (g *Game) MakeMove(m Move) bool {
	if g.IsOver() {
		return false
	}
	legal := g.LegalMoves()
	for _, lm := range legal {
		if lm.From == m.From && lm.To == m.To && lm.Promotion == m.Promotion {
			pos := g.Current().ApplyMove(lm)
			g.History = append(g.History, pos)
			g.Moves = append(g.Moves, lm)
			g.updateStatus()
			return true
		}
	}
	return false
}

func (g *Game) UndoMove() bool {
	if len(g.History) <= 1 {
		return false
	}
	g.History = g.History[:len(g.History)-1]
	if len(g.Moves) > 0 {
		g.Moves = g.Moves[:len(g.Moves)-1]
	}
	g.updateStatus()
	return true
}

func (g *Game) updateStatus() {
	pos := g.Current()
	legal := pos.GenerateLegalMoves()
	inCheck := pos.IsInCheck(pos.ActiveColor)

	if len(legal) == 0 {
		if inCheck {
			g.Status = Checkmate
		} else {
			g.Status = Stalemate
		}
		return
	}
	if pos.HalfMoveClock >= 100 {
		g.Status = Draw50Move
		return
	}
	if g.isInsufficientMaterial() {
		g.Status = DrawInsufficient
		return
	}
	if inCheck {
		g.Status = InCheck
	} else {
		g.Status = Playing
	}
}

func (g *Game) isInsufficientMaterial() bool {
	pos := g.Current()
	var pieces []Piece
	for sq := Square(0); sq < 64; sq++ {
		p := pos.Board[sq]
		if !p.IsEmpty() {
			pieces = append(pieces, p)
		}
	}
	if len(pieces) == 2 {
		return true
	}
	if len(pieces) == 3 {
		for _, p := range pieces {
			if p.Type == Knight || p.Type == Bishop {
				return true
			}
		}
	}
	return false
}

func (g *Game) IsOver() bool {
	return g.Status == Checkmate || g.Status == Stalemate ||
		g.Status == Draw50Move || g.Status == DrawInsufficient
}

func (g *Game) Winner() Color {
	if g.Status == Checkmate {
		return g.Current().ActiveColor.Opposite()
	}
	return NoColor
}
