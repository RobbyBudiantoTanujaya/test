package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/robbybt94/test/chess"
)

var (
	mu   sync.Mutex
	game = chess.NewGame()
)

type PieceInfo struct {
	Type  string `json:"type"`
	Color string `json:"color"`
}

type StateResponse struct {
	Board       [8][8]*PieceInfo    `json:"board"`
	ActiveColor string              `json:"activeColor"`
	Status      string              `json:"status"`
	LegalMoves  map[string][]string `json:"legalMoves"`
	Moves       []string            `json:"moves"`
	Winner      string              `json:"winner"`
}

func pieceTypeStr(pt chess.PieceType) string {
	switch pt {
	case chess.Pawn:
		return "pawn"
	case chess.Knight:
		return "knight"
	case chess.Bishop:
		return "bishop"
	case chess.Rook:
		return "rook"
	case chess.Queen:
		return "queen"
	case chess.King:
		return "king"
	}
	return ""
}

func buildState() StateResponse {
	pos := game.Current()
	var board [8][8]*PieceInfo
	for sq := chess.Square(0); sq < 64; sq++ {
		p := pos.Board[sq]
		if !p.IsEmpty() {
			board[sq.Row()][sq.Col()] = &PieceInfo{
				Type:  pieceTypeStr(p.Type),
				Color: p.Color.String(),
			}
		}
	}

	legal := game.LegalMoves()
	legalMap := make(map[string][]string)
	for _, m := range legal {
		from := m.From.String()
		to := m.To.String()
		if m.Promotion != chess.NoPiece {
			switch m.Promotion {
			case chess.Queen:
				to += "q"
			case chess.Rook:
				to += "r"
			case chess.Bishop:
				to += "b"
			case chess.Knight:
				to += "n"
			}
		}
		legalMap[from] = append(legalMap[from], to)
	}

	var moveList []string
	for _, m := range game.Moves {
		moveList = append(moveList, m.String())
	}

	winner := ""
	if game.Status == chess.Checkmate {
		winner = game.Winner().String()
	}

	return StateResponse{
		Board:       board,
		ActiveColor: pos.ActiveColor.String(),
		Status:      game.Status.String(),
		LegalMoves:  legalMap,
		Moves:       moveList,
		Winner:      winner,
	}
}

func jsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("static")))

	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		jsonResponse(w, buildState())
	})

	mux.HandleFunc("POST /api/new", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		game = chess.NewGame()
		jsonResponse(w, buildState())
	})

	mux.HandleFunc("POST /api/move", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		var req struct {
			From      string `json:"from"`
			To        string `json:"to"`
			Promotion string `json:"promotion"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		from := chess.ParseSquare(req.From)
		to := chess.ParseSquare(req.To)
		if !from.Valid() || !to.Valid() {
			http.Error(w, "invalid square", http.StatusBadRequest)
			return
		}

		var promo chess.PieceType
		switch req.Promotion {
		case "q":
			promo = chess.Queen
		case "r":
			promo = chess.Rook
		case "b":
			promo = chess.Bishop
		case "n":
			promo = chess.Knight
		}

		if !game.MakeMove(chess.Move{From: from, To: to, Promotion: promo}) {
			http.Error(w, "illegal move", http.StatusBadRequest)
			return
		}
		jsonResponse(w, buildState())
	})

	mux.HandleFunc("POST /api/ai", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if game.IsOver() {
			http.Error(w, "game over", http.StatusBadRequest)
			return
		}
		m := chess.BestMove(game.Current(), 4)
		if m == (chess.Move{}) {
			http.Error(w, "no moves available", http.StatusInternalServerError)
			return
		}
		game.MakeMove(m)
		jsonResponse(w, buildState())
	})

	mux.HandleFunc("POST /api/undo", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		game.UndoMove()
		jsonResponse(w, buildState())
	})

	log.Println("Chess server → http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
