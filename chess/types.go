package chess

type Color int

const (
	White Color = iota
	Black
	NoColor
)

func (c Color) Opposite() Color {
	if c == White {
		return Black
	}
	return White
}

func (c Color) String() string {
	switch c {
	case White:
		return "white"
	case Black:
		return "black"
	}
	return ""
}

type PieceType int

const (
	NoPiece PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

type Piece struct {
	Type  PieceType
	Color Color
}

var EmptyPiece = Piece{}

func (p Piece) IsEmpty() bool { return p.Type == NoPiece }

type Square int

const InvalidSquare Square = -1

func Sq(row, col int) Square {
	if row < 0 || row > 7 || col < 0 || col > 7 {
		return InvalidSquare
	}
	return Square(row*8 + col)
}

func (s Square) Row() int    { return int(s) / 8 }
func (s Square) Col() int    { return int(s) % 8 }
func (s Square) Valid() bool { return s >= 0 && s <= 63 }
func (s Square) File() byte  { return byte('a' + s.Col()) }
func (s Square) Rank() byte  { return byte('1' + s.Row()) }

func (s Square) String() string {
	if !s.Valid() {
		return "-"
	}
	return string([]byte{s.File(), s.Rank()})
}

func ParseSquare(str string) Square {
	if len(str) != 2 {
		return InvalidSquare
	}
	col := int(str[0] - 'a')
	row := int(str[1] - '1')
	return Sq(row, col)
}

type Move struct {
	From      Square
	To        Square
	Promotion PieceType
	Castle    bool
	EnPassant bool
}

func (m Move) String() string {
	s := m.From.String() + m.To.String()
	switch m.Promotion {
	case Queen:
		s += "q"
	case Rook:
		s += "r"
	case Bishop:
		s += "b"
	case Knight:
		s += "n"
	}
	return s
}
