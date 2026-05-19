package chess

type CastleRights struct {
	WhiteKingside  bool
	WhiteQueenside bool
	BlackKingside  bool
	BlackQueenside bool
}

type Position struct {
	Board         [64]Piece
	ActiveColor   Color
	Castle        CastleRights
	EnPassant     Square
	HalfMoveClock int
	FullMove      int
}

func NewStartPosition() *Position {
	p := &Position{
		ActiveColor:   White,
		Castle:        CastleRights{true, true, true, true},
		EnPassant:     InvalidSquare,
		HalfMoveClock: 0,
		FullMove:      1,
	}
	backRank := []PieceType{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}
	for col, pt := range backRank {
		p.Board[Sq(0, col)] = Piece{pt, White}
		p.Board[Sq(7, col)] = Piece{pt, Black}
	}
	for col := 0; col < 8; col++ {
		p.Board[Sq(1, col)] = Piece{Pawn, White}
		p.Board[Sq(6, col)] = Piece{Pawn, Black}
	}
	return p
}

func (p *Position) FindKing(color Color) Square {
	for sq := Square(0); sq < 64; sq++ {
		piece := p.Board[sq]
		if piece.Type == King && piece.Color == color {
			return sq
		}
	}
	return InvalidSquare
}

func (p *Position) ApplyMove(m Move) *Position {
	next := *p
	next.EnPassant = InvalidSquare

	piece := p.Board[m.From]

	if piece.Type == King {
		if piece.Color == White {
			next.Castle.WhiteKingside = false
			next.Castle.WhiteQueenside = false
		} else {
			next.Castle.BlackKingside = false
			next.Castle.BlackQueenside = false
		}
	}
	if piece.Type == Rook {
		switch m.From {
		case Sq(0, 0):
			next.Castle.WhiteQueenside = false
		case Sq(0, 7):
			next.Castle.WhiteKingside = false
		case Sq(7, 0):
			next.Castle.BlackQueenside = false
		case Sq(7, 7):
			next.Castle.BlackKingside = false
		}
	}
	captured := p.Board[m.To]
	if captured.Type == Rook {
		switch m.To {
		case Sq(0, 0):
			next.Castle.WhiteQueenside = false
		case Sq(0, 7):
			next.Castle.WhiteKingside = false
		case Sq(7, 0):
			next.Castle.BlackQueenside = false
		case Sq(7, 7):
			next.Castle.BlackKingside = false
		}
	}

	next.Board[m.To] = piece
	next.Board[m.From] = EmptyPiece

	if m.EnPassant {
		capturedSq := Sq(m.From.Row(), m.To.Col())
		next.Board[capturedSq] = EmptyPiece
	}

	if m.Castle {
		if m.To.Col() == 6 {
			rookFrom := Sq(m.From.Row(), 7)
			rookTo := Sq(m.From.Row(), 5)
			next.Board[rookTo] = next.Board[rookFrom]
			next.Board[rookFrom] = EmptyPiece
		} else {
			rookFrom := Sq(m.From.Row(), 0)
			rookTo := Sq(m.From.Row(), 3)
			next.Board[rookTo] = next.Board[rookFrom]
			next.Board[rookFrom] = EmptyPiece
		}
	}

	if m.Promotion != NoPiece {
		next.Board[m.To] = Piece{m.Promotion, piece.Color}
	}

	if piece.Type == Pawn && iabs(m.From.Row()-m.To.Row()) == 2 {
		epRow := (m.From.Row() + m.To.Row()) / 2
		next.EnPassant = Sq(epRow, m.From.Col())
	}

	if piece.Type == Pawn || !captured.IsEmpty() {
		next.HalfMoveClock = 0
	} else {
		next.HalfMoveClock++
	}

	if p.ActiveColor == Black {
		next.FullMove++
	}
	next.ActiveColor = p.ActiveColor.Opposite()

	return &next
}

func iabs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
