package chess

func (p *Position) GeneratePseudoMoves() []Move {
	var moves []Move
	for sq := Square(0); sq < 64; sq++ {
		piece := p.Board[sq]
		if piece.IsEmpty() || piece.Color != p.ActiveColor {
			continue
		}
		moves = append(moves, p.pieceMoves(sq, piece)...)
	}
	return moves
}

func (p *Position) GenerateLegalMoves() []Move {
	pseudo := p.GeneratePseudoMoves()
	legal := pseudo[:0:len(pseudo)]
	for _, m := range pseudo {
		next := p.ApplyMove(m)
		if !next.IsInCheck(p.ActiveColor) {
			legal = append(legal, m)
		}
	}
	return legal
}

func (p *Position) pieceMoves(sq Square, piece Piece) []Move {
	switch piece.Type {
	case Pawn:
		return p.pawnMoves(sq, piece.Color)
	case Knight:
		return p.knightMoves(sq, piece.Color)
	case Bishop:
		return p.slidingMoves(sq, piece.Color, [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}})
	case Rook:
		return p.slidingMoves(sq, piece.Color, [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}})
	case Queen:
		return p.slidingMoves(sq, piece.Color, [][2]int{
			{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
			{0, 1}, {0, -1}, {1, 0}, {-1, 0},
		})
	case King:
		return p.kingMoves(sq, piece.Color)
	}
	return nil
}

func (p *Position) pawnMoves(sq Square, color Color) []Move {
	var moves []Move
	row, col := sq.Row(), sq.Col()

	dir := 1
	startRow := 1
	promoRow := 7
	if color == Black {
		dir = -1
		startRow = 6
		promoRow = 0
	}

	fwd := Sq(row+dir, col)
	if fwd.Valid() && p.Board[fwd].IsEmpty() {
		if row+dir == promoRow {
			for _, pt := range []PieceType{Queen, Rook, Bishop, Knight} {
				moves = append(moves, Move{From: sq, To: fwd, Promotion: pt})
			}
		} else {
			moves = append(moves, Move{From: sq, To: fwd})
			if row == startRow {
				fwd2 := Sq(row+2*dir, col)
				if fwd2.Valid() && p.Board[fwd2].IsEmpty() {
					moves = append(moves, Move{From: sq, To: fwd2})
				}
			}
		}
	}

	for _, dcol := range []int{-1, 1} {
		diag := Sq(row+dir, col+dcol)
		if !diag.Valid() {
			continue
		}
		target := p.Board[diag]
		if !target.IsEmpty() && target.Color != color {
			if row+dir == promoRow {
				for _, pt := range []PieceType{Queen, Rook, Bishop, Knight} {
					moves = append(moves, Move{From: sq, To: diag, Promotion: pt})
				}
			} else {
				moves = append(moves, Move{From: sq, To: diag})
			}
		}
		if p.EnPassant == diag {
			moves = append(moves, Move{From: sq, To: diag, EnPassant: true})
		}
	}
	return moves
}

func (p *Position) knightMoves(sq Square, color Color) []Move {
	var moves []Move
	row, col := sq.Row(), sq.Col()
	for _, off := range [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}} {
		dest := Sq(row+off[0], col+off[1])
		if dest.Valid() {
			t := p.Board[dest]
			if t.IsEmpty() || t.Color != color {
				moves = append(moves, Move{From: sq, To: dest})
			}
		}
	}
	return moves
}

func (p *Position) slidingMoves(sq Square, color Color, dirs [][2]int) []Move {
	var moves []Move
	row, col := sq.Row(), sq.Col()
	for _, dir := range dirs {
		for i := 1; i < 8; i++ {
			dest := Sq(row+dir[0]*i, col+dir[1]*i)
			if !dest.Valid() {
				break
			}
			t := p.Board[dest]
			if t.IsEmpty() {
				moves = append(moves, Move{From: sq, To: dest})
			} else if t.Color != color {
				moves = append(moves, Move{From: sq, To: dest})
				break
			} else {
				break
			}
		}
	}
	return moves
}

func (p *Position) kingMoves(sq Square, color Color) []Move {
	var moves []Move
	row, col := sq.Row(), sq.Col()
	for _, off := range [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}} {
		dest := Sq(row+off[0], col+off[1])
		if dest.Valid() {
			t := p.Board[dest]
			if t.IsEmpty() || t.Color != color {
				moves = append(moves, Move{From: sq, To: dest})
			}
		}
	}

	opponent := color.Opposite()
	if color == White && row == 0 {
		if p.Castle.WhiteKingside &&
			p.Board[Sq(0, 5)].IsEmpty() &&
			p.Board[Sq(0, 6)].IsEmpty() &&
			!p.IsSquareAttacked(Sq(0, 4), opponent) &&
			!p.IsSquareAttacked(Sq(0, 5), opponent) &&
			!p.IsSquareAttacked(Sq(0, 6), opponent) {
			moves = append(moves, Move{From: sq, To: Sq(0, 6), Castle: true})
		}
		if p.Castle.WhiteQueenside &&
			p.Board[Sq(0, 3)].IsEmpty() &&
			p.Board[Sq(0, 2)].IsEmpty() &&
			p.Board[Sq(0, 1)].IsEmpty() &&
			!p.IsSquareAttacked(Sq(0, 4), opponent) &&
			!p.IsSquareAttacked(Sq(0, 3), opponent) &&
			!p.IsSquareAttacked(Sq(0, 2), opponent) {
			moves = append(moves, Move{From: sq, To: Sq(0, 2), Castle: true})
		}
	} else if color == Black && row == 7 {
		if p.Castle.BlackKingside &&
			p.Board[Sq(7, 5)].IsEmpty() &&
			p.Board[Sq(7, 6)].IsEmpty() &&
			!p.IsSquareAttacked(Sq(7, 4), opponent) &&
			!p.IsSquareAttacked(Sq(7, 5), opponent) &&
			!p.IsSquareAttacked(Sq(7, 6), opponent) {
			moves = append(moves, Move{From: sq, To: Sq(7, 6), Castle: true})
		}
		if p.Castle.BlackQueenside &&
			p.Board[Sq(7, 3)].IsEmpty() &&
			p.Board[Sq(7, 2)].IsEmpty() &&
			p.Board[Sq(7, 1)].IsEmpty() &&
			!p.IsSquareAttacked(Sq(7, 4), opponent) &&
			!p.IsSquareAttacked(Sq(7, 3), opponent) &&
			!p.IsSquareAttacked(Sq(7, 2), opponent) {
			moves = append(moves, Move{From: sq, To: Sq(7, 2), Castle: true})
		}
	}
	return moves
}

func (p *Position) IsSquareAttacked(sq Square, byColor Color) bool {
	row, col := sq.Row(), sq.Col()

	pawnDir := 1
	if byColor == White {
		pawnDir = -1
	}
	for _, dcol := range []int{-1, 1} {
		a := Sq(row+pawnDir, col+dcol)
		if a.Valid() {
			ap := p.Board[a]
			if ap.Type == Pawn && ap.Color == byColor {
				return true
			}
		}
	}

	for _, off := range [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}} {
		a := Sq(row+off[0], col+off[1])
		if a.Valid() {
			ap := p.Board[a]
			if ap.Type == Knight && ap.Color == byColor {
				return true
			}
		}
	}

	for _, dir := range [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		for i := 1; i < 8; i++ {
			a := Sq(row+dir[0]*i, col+dir[1]*i)
			if !a.Valid() {
				break
			}
			ap := p.Board[a]
			if !ap.IsEmpty() {
				if ap.Color == byColor && (ap.Type == Bishop || ap.Type == Queen) {
					return true
				}
				break
			}
		}
	}

	for _, dir := range [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}} {
		for i := 1; i < 8; i++ {
			a := Sq(row+dir[0]*i, col+dir[1]*i)
			if !a.Valid() {
				break
			}
			ap := p.Board[a]
			if !ap.IsEmpty() {
				if ap.Color == byColor && (ap.Type == Rook || ap.Type == Queen) {
					return true
				}
				break
			}
		}
	}

	for _, off := range [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}} {
		a := Sq(row+off[0], col+off[1])
		if a.Valid() {
			ap := p.Board[a]
			if ap.Type == King && ap.Color == byColor {
				return true
			}
		}
	}

	return false
}

func (p *Position) IsInCheck(color Color) bool {
	king := p.FindKing(color)
	if !king.Valid() {
		return false
	}
	return p.IsSquareAttacked(king, color.Opposite())
}
