'use strict';

// ── Unicode piece glyphs ──────────────────────────────────────────────────────
const GLYPHS = {
  white: { king: '♔', queen: '♕', rook: '♖', bishop: '♗', knight: '♘', pawn: '♙' },
  black: { king: '♚', queen: '♛', rook: '♜', bishop: '♝', knight: '♞', pawn: '♟' },
};

// ── State ─────────────────────────────────────────────────────────────────────
let state = null;          // last server response
let selected = null;       // currently selected square string e.g. "e2"
let pendingPromo = null;   // { from, to } awaiting promotion choice
let flipped = false;       // board orientation

// ── DOM refs ──────────────────────────────────────────────────────────────────
const boardEl     = document.getElementById('board');
const statusText  = document.getElementById('status-text');
const turnIndEl   = document.getElementById('turn-indicator');
const moveListEl  = document.getElementById('move-list');
const btnNew      = document.getElementById('btn-new');
const btnUndo     = document.getElementById('btn-undo');
const promoOverlay = document.getElementById('promotion-overlay');
const promoChoices = document.getElementById('promotion-choices');

// ── Coordinate labels ─────────────────────────────────────────────────────────
function buildCoords() {
  const files = ['a','b','c','d','e','f','g','h'];
  const ranks = ['1','2','3','4','5','6','7','8'];

  const makeSpan = (t) => { const s = document.createElement('span'); s.className = 'coord'; s.textContent = t; return s; };

  const top    = document.getElementById('coords-top');
  const bottom = document.getElementById('coords-bottom');
  const left   = document.getElementById('coords-left');
  const right  = document.getElementById('coords-right');

  [top, bottom].forEach(el => {
    el.innerHTML = '';
    (flipped ? [...files].reverse() : files).forEach(f => el.appendChild(makeSpan(f)));
  });

  const displayRanks = flipped ? ranks : [...ranks].reverse();
  [left, right].forEach(el => {
    el.innerHTML = '';
    displayRanks.forEach(r => el.appendChild(makeSpan(r)));
  });
}

// ── Convert square string to row/col from display perspective ─────────────────
function sqToDisplay(sq) {
  const col = sq.charCodeAt(0) - 97; // a=0
  const row = parseInt(sq[1]) - 1;
  return { row, col };
}

function displayToSq(dRow, dCol) {
  const col = flipped ? 7 - dCol : dCol;
  const row = flipped ? dRow : 7 - dRow;
  return String.fromCharCode(97 + col) + (row + 1);
}

// ── Main render ───────────────────────────────────────────────────────────────
function render() {
  if (!state) return;
  buildCoords();
  renderBoard();
  renderStatus();
  renderMoveList();
}

function renderBoard() {
  boardEl.innerHTML = '';

  const lastFrom = state.moves && state.moves.length > 0 ? state.moves[state.moves.length - 1].slice(0, 2) : null;
  const lastTo   = state.moves && state.moves.length > 0 ? state.moves[state.moves.length - 1].slice(2, 4) : null;

  // Find king in check
  let checkKingSq = null;
  if (state.status === 'check' || state.status === 'checkmate') {
    const color = state.activeColor;
    for (let r = 0; r < 8; r++) {
      for (let c = 0; c < 8; c++) {
        const p = state.board[r][c];
        if (p && p.type === 'king' && p.color === color) {
          checkKingSq = String.fromCharCode(97 + c) + (r + 1);
        }
      }
    }
  }

  // Valid destinations for selected piece
  const validDests = new Set();
  if (selected && state.legalMoves[selected]) {
    state.legalMoves[selected].forEach(to => validDests.add(to.slice(0, 2)));
  }

  for (let dRow = 0; dRow < 8; dRow++) {
    for (let dCol = 0; dCol < 8; dCol++) {
      const sq = displayToSq(dRow, dCol);
      const { row, col } = sqToDisplay(sq);

      const sqEl = document.createElement('div');
      sqEl.className = 'square ' + ((row + col) % 2 === 0 ? 'light' : 'dark');
      sqEl.dataset.sq = sq;

      if (sq === selected)  sqEl.classList.add('selected');
      if (sq === lastFrom)  sqEl.classList.add('last-from');
      if (sq === lastTo)    sqEl.classList.add('last-to');
      if (sq === checkKingSq) sqEl.classList.add('in-check');

      const piece = state.board[row][col];
      if (piece) {
        const pieceEl = document.createElement('span');
        pieceEl.className = 'piece';
        pieceEl.textContent = GLYPHS[piece.color][piece.type];
        sqEl.appendChild(pieceEl);
      }

      // Move hints
      if (selected && validDests.has(sq)) {
        const hint = document.createElement('div');
        hint.className = piece ? 'hint-capture' : 'hint-dot';
        sqEl.appendChild(hint);
      }

      sqEl.addEventListener('click', () => onSquareClick(sq));
      boardEl.appendChild(sqEl);
    }
  }
}

function renderStatus() {
  const color = state.activeColor;
  const colorLabel = color.charAt(0).toUpperCase() + color.slice(1);

  turnIndEl.textContent = state.status === 'checkmate' || state.status === 'stalemate'
    ? 'Game Over'
    : `${colorLabel}'s turn`;

  const statusMap = {
    playing:          '',
    check:            '⚠ Check!',
    checkmate:        `♟ Checkmate — ${state.winner.charAt(0).toUpperCase() + state.winner.slice(1)} wins!`,
    stalemate:        '½ Stalemate — Draw',
    draw50:           '½ Draw (50-move rule)',
    drawInsufficient: '½ Draw (insufficient material)',
  };
  statusText.textContent = statusMap[state.status] ?? '';
}

function renderMoveList() {
  moveListEl.innerHTML = '';
  const moves = state.moves || [];
  for (let i = 0; i < moves.length; i += 2) {
    const numEl = document.createElement('span');
    numEl.className = 'move-num';
    numEl.textContent = (i / 2 + 1) + '.';

    const w = document.createElement('span');
    w.className = 'move-cell';
    w.textContent = moves[i];

    const b = document.createElement('span');
    b.className = 'move-cell';
    b.textContent = moves[i + 1] || '';

    moveListEl.appendChild(numEl);
    moveListEl.appendChild(w);
    moveListEl.appendChild(b);
  }
  moveListEl.scrollTop = moveListEl.scrollHeight;
}

// ── Square click handler ───────────────────────────────────────────────────────
function onSquareClick(sq) {
  if (!state || state.status === 'checkmate' || state.status === 'stalemate' ||
      state.status === 'draw50' || state.status === 'drawInsufficient') return;

  const { row, col } = sqToDisplay(sq);
  const piece = state.board[row][col];

  // If a piece is already selected
  if (selected) {
    if (sq === selected) {
      selected = null;
      render();
      return;
    }

    const validTos = state.legalMoves[selected] || [];
    const matchingMoves = validTos.filter(to => to.slice(0, 2) === sq);

    if (matchingMoves.length > 0) {
      // Check if promotion needed
      const promoMoves = matchingMoves.filter(to => to.length > 2);
      if (promoMoves.length > 0) {
        pendingPromo = { from: selected, to: sq };
        showPromotion(state.activeColor);
        return;
      }
      executeMove(selected, sq, '');
      return;
    }

    // Clicked on own piece — switch selection
    if (piece && piece.color === state.activeColor) {
      selected = sq;
      render();
      return;
    }

    selected = null;
    render();
    return;
  }

  // No selection yet
  if (piece && piece.color === state.activeColor) {
    selected = sq;
    render();
  }
}

// ── Move execution ────────────────────────────────────────────────────────────
async function executeMove(from, to, promotion) {
  selected = null;
  try {
    const res = await fetch('/api/move', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ from, to, promotion }),
    });
    if (!res.ok) return;
    state = await res.json();
    render();

    // Auto AI move if needed
    const mode = document.querySelector('input[name="mode"]:checked').value;
    if (!isGameOver() && shouldAiMove(mode)) {
      setTimeout(triggerAI, 200);
    }
  } catch (e) {
    console.error(e);
  }
}

function isGameOver() {
  return ['checkmate','stalemate','draw50','drawInsufficient'].includes(state.status);
}

function shouldAiMove(mode) {
  if (mode === '2p') return false;
  if (mode === 'white') return state.activeColor === 'black';
  if (mode === 'black') return state.activeColor === 'white';
  return false;
}

async function triggerAI() {
  if (!state || isGameOver()) return;
  turnIndEl.innerHTML = `${state.activeColor.charAt(0).toUpperCase() + state.activeColor.slice(1)} thinking<span class="thinking"></span>`;
  try {
    const res = await fetch('/api/ai', { method: 'POST' });
    if (!res.ok) return;
    state = await res.json();
    render();
  } catch (e) {
    console.error(e);
  }
}

// ── Promotion dialog ──────────────────────────────────────────────────────────
function showPromotion(color) {
  promoChoices.innerHTML = '';
  const pieces = [
    { type: 'queen', code: 'q' },
    { type: 'rook',  code: 'r' },
    { type: 'bishop',code: 'b' },
    { type: 'knight',code: 'n' },
  ];
  pieces.forEach(({ type, code }) => {
    const btn = document.createElement('div');
    btn.className = 'promo-btn';
    btn.textContent = GLYPHS[color][type];
    btn.addEventListener('click', () => {
      promoOverlay.classList.add('hidden');
      const { from, to } = pendingPromo;
      pendingPromo = null;
      executeMove(from, to, code);
    });
    promoChoices.appendChild(btn);
  });
  promoOverlay.classList.remove('hidden');
}

// ── Controls ──────────────────────────────────────────────────────────────────
btnNew.addEventListener('click', async () => {
  selected = null;
  const res = await fetch('/api/new', { method: 'POST' });
  state = await res.json();
  render();
  // If playing as black, trigger AI first move
  const mode = document.querySelector('input[name="mode"]:checked').value;
  if (mode === 'black') setTimeout(triggerAI, 200);
});

btnUndo.addEventListener('click', async () => {
  selected = null;
  const mode = document.querySelector('input[name="mode"]:checked').value;
  // Undo twice in AI mode (undo AI response and player move)
  const res = await fetch('/api/undo', { method: 'POST' });
  state = await res.json();
  if (mode !== '2p') {
    const res2 = await fetch('/api/undo', { method: 'POST' });
    state = await res2.json();
  }
  render();
});

// Mode change — flip board for black
document.querySelectorAll('input[name="mode"]').forEach(radio => {
  radio.addEventListener('change', () => {
    flipped = radio.value === 'black';
    render();
  });
});

// ── Init ──────────────────────────────────────────────────────────────────────
(async () => {
  const res = await fetch('/api/state');
  state = await res.json();
  render();
})();
