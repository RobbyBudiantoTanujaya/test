package chess

const infinity = 1_000_000

var pieceValues = map[PieceType]int{
	Pawn:   100,
	Knight: 320,
	Bishop: 330,
	Rook:   500,
	Queen:  900,
	King:   20000,
}

// Piece-square tables (from white's perspective, row 0 = rank 1)
var pawnPST = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, -20, -20, 10, 10, 5,
	5, -5, -10, 0, 0, -10, -5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, 5, 10, 25, 25, 10, 5, 5,
	10, 10, 20, 30, 30, 20, 10, 10,
	50, 50, 50, 50, 50, 50, 50, 50,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightPST = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopPST = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookPST = [64]int{
	0, 0, 0, 5, 5, 0, 0, 0,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	5, 10, 10, 10, 10, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var queenPST = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-10, 5, 5, 5, 5, 5, 0, -10,
	0, 0, 5, 5, 5, 5, 0, -5,
	-5, 0, 5, 5, 5, 5, 0, -5,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingPST = [64]int{
	20, 30, 10, 0, 0, 10, 30, 20,
	20, 20, 0, 0, 0, 0, 20, 20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
}

func pstValue(pt PieceType, sq Square, color Color) int {
	var pst *[64]int
	switch pt {
	case Pawn:
		pst = &pawnPST
	case Knight:
		pst = &knightPST
	case Bishop:
		pst = &bishopPST
	case Rook:
		pst = &rookPST
	case Queen:
		pst = &queenPST
	case King:
		pst = &kingPST
	default:
		return 0
	}
	if color == White {
		return pst[sq]
	}
	mirror := Square((7-sq.Row())*8 + sq.Col())
	return pst[mirror]
}

func Evaluate(pos *Position) int {
	score := 0
	for sq := Square(0); sq < 64; sq++ {
		p := pos.Board[sq]
		if p.IsEmpty() {
			continue
		}
		v := pieceValues[p.Type] + pstValue(p.Type, sq, p.Color)
		if p.Color == White {
			score += v
		} else {
			score -= v
		}
	}
	return score
}

func search(pos *Position, depth, alpha, beta int, maximizing bool) int {
	if depth == 0 {
		return Evaluate(pos)
	}
	moves := pos.GenerateLegalMoves()
	if len(moves) == 0 {
		if pos.IsInCheck(pos.ActiveColor) {
			if maximizing {
				return -infinity + depth
			}
			return infinity - depth
		}
		return 0
	}

	if maximizing {
		best := -infinity
		for _, m := range moves {
			score := search(pos.ApplyMove(m), depth-1, alpha, beta, false)
			if score > best {
				best = score
			}
			if best > alpha {
				alpha = best
			}
			if beta <= alpha {
				break
			}
		}
		return best
	}
	best := infinity
	for _, m := range moves {
		score := search(pos.ApplyMove(m), depth-1, alpha, beta, true)
		if score < best {
			best = score
		}
		if best < beta {
			beta = best
		}
		if beta <= alpha {
			break
		}
	}
	return best
}

func BestMove(pos *Position, depth int) Move {
	moves := pos.GenerateLegalMoves()
	if len(moves) == 0 {
		return Move{}
	}

	maximizing := pos.ActiveColor == White
	bestScore := -infinity
	if !maximizing {
		bestScore = infinity
	}
	best := moves[0]
	alpha, beta := -infinity, infinity

	for _, m := range moves {
		score := search(pos.ApplyMove(m), depth-1, alpha, beta, !maximizing)
		if maximizing && score > bestScore {
			bestScore = score
			best = m
			if bestScore > alpha {
				alpha = bestScore
			}
		} else if !maximizing && score < bestScore {
			bestScore = score
			best = m
			if bestScore < beta {
				beta = bestScore
			}
		}
	}
	return best
}
