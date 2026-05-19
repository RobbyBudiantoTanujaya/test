'use strict';

// ── Piece asset mapping ───────────────────────────────────────────────────────
// Files: static/pieces/{color_initial}{Type_initial}.svg
// e.g. white king → wK.svg, black pawn → bP.svg
const TYPE_CHAR = { king:'K', queen:'Q', rook:'R', bishop:'B', knight:'N', pawn:'P' };

function pieceImg(color, type) {
  const img = document.createElement('img');
  img.src = `/pieces/${color[0]}${TYPE_CHAR[type]}.svg`;
  img.className = 'piece';
  img.alt = `${color} ${type}`;
  img.draggable = false;
  return img;
}

// ── State ─────────────────────────────────────────────────────────────────────
let state       = null;
let selected    = null;   // selected square string e.g. "e2"
let pendingPromo = null;  // { from, to }
let flipped     = false;  // board orientation (true when playing as black)

// ── DOM ───────────────────────────────────────────────────────────────────────
const boardEl       = document.getElementById('board');
const statusText    = document.getElementById('status-text');
const turnIndEl     = document.getElementById('turn-indicator');
const moveListEl    = document.getElementById('move-list');
const btnNew        = document.getElementById('btn-new');
const btnUndo       = document.getElementById('btn-undo');
const promoOverlay  = document.getElementById('promotion-overlay');
const promoChoices  = document.getElementById('promotion-choices');

// ── Coordinate helpers ────────────────────────────────────────────────────────
// Given a display cell (dRow=0 is top of screen), return chess square string.
function displayToSq(dRow, dCol) {
  const col = flipped ? 7 - dCol : dCol;
  const row = flipped ? dRow     : 7 - dRow;
  return String.fromCharCode(97 + col) + (row + 1);
}

// ── Render ────────────────────────────────────────────────────────────────────
function render() {
  if (!state) return;
  renderBoard();
  renderStatus();
  renderMoveList();
}

function renderBoard() {
  boardEl.innerHTML = '';

  const lastMove = state.moves?.length ? state.moves[state.moves.length - 1] : null;
  const lastFrom = lastMove ? lastMove.slice(0, 2) : null;
  const lastTo   = lastMove ? lastMove.slice(2, 4) : null;

  // King in check square
  let checkKingSq = null;
  if (state.status === 'check' || state.status === 'checkmate') {
    const color = state.activeColor;
    outer: for (let r = 0; r < 8; r++) {
      for (let c = 0; c < 8; c++) {
        const p = state.board[r][c];
        if (p?.type === 'king' && p.color === color) {
          checkKingSq = String.fromCharCode(97 + c) + (r + 1);
          break outer;
        }
      }
    }
  }

  // Valid destinations for selected piece
  const validDests = new Set();
  if (selected && state.legalMoves[selected]) {
    state.legalMoves[selected].forEach(to => validDests.add(to.slice(0, 2)));
  }

  // Files and ranks for coordinate labels
  const files = flipped ? ['h','g','f','e','d','c','b','a'] : ['a','b','c','d','e','f','g','h'];
  const ranks = flipped ? ['1','2','3','4','5','6','7','8'] : ['8','7','6','5','4','3','2','1'];

  for (let dRow = 0; dRow < 8; dRow++) {
    for (let dCol = 0; dCol < 8; dCol++) {
      const sq = displayToSq(dRow, dCol);
      const col = sq.charCodeAt(0) - 97;
      const row = parseInt(sq[1]) - 1;

      const sqEl = document.createElement('div');
      const isLight = (row + col) % 2 === 0;
      sqEl.className = 'square ' + (isLight ? 'light' : 'dark');
      sqEl.dataset.sq = sq;

      // Apply highlights
      if (sq === selected)    sqEl.classList.add('selected');
      if (sq === lastFrom)    sqEl.classList.add('last-from');
      if (sq === lastTo)      sqEl.classList.add('last-to');
      if (sq === checkKingSq) sqEl.classList.add('in-check');

      // ── Coordinate labels ──
      // Rank label: left edge column (dCol === 0)
      if (dCol === 0) {
        const rankEl = document.createElement('span');
        rankEl.className = 'coord-label coord-rank';
        rankEl.textContent = ranks[dRow];
        sqEl.appendChild(rankEl);
      }
      // File label: bottom row (dRow === 7)
      if (dRow === 7) {
        const fileEl = document.createElement('span');
        fileEl.className = 'coord-label coord-file';
        fileEl.textContent = files[dCol];
        sqEl.appendChild(fileEl);
      }

      // ── Piece ──
      const piece = state.board[row][col];
      if (piece) {
        sqEl.appendChild(pieceImg(piece.color, piece.type));
      }

      // ── Move hints ──
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
  const cap   = s => s.charAt(0).toUpperCase() + s.slice(1);
  const isOver = ['checkmate','stalemate','draw50','drawInsufficient'].includes(state.status);

  turnIndEl.textContent = isOver ? 'Game Over' : `${cap(color)}'s turn`;

  const msgs = {
    playing:          '',
    check:            '⚠ Check!',
    checkmate:        `♟ Checkmate — ${cap(state.winner)} wins!`,
    stalemate:        '½ Stalemate — Draw',
    draw50:           '½ Draw (50-move rule)',
    drawInsufficient: '½ Draw (insufficient material)',
  };
  statusText.textContent = msgs[state.status] ?? '';
}

function renderMoveList() {
  moveListEl.innerHTML = '';
  const moves = state.moves || [];
  for (let i = 0; i < moves.length; i += 2) {
    const numEl = document.createElement('span');
    numEl.className = 'move-num';
    numEl.textContent = (i / 2 + 1) + '.';

    const wEl = document.createElement('span');
    wEl.className = 'move-cell';
    wEl.textContent = moves[i];

    const bEl = document.createElement('span');
    bEl.className = 'move-cell';
    bEl.textContent = moves[i + 1] || '';

    moveListEl.appendChild(numEl);
    moveListEl.appendChild(wEl);
    moveListEl.appendChild(bEl);
  }
  moveListEl.scrollTop = moveListEl.scrollHeight;
}

// ── Click handler ─────────────────────────────────────────────────────────────
function onSquareClick(sq) {
  if (!state) return;
  if (['checkmate','stalemate','draw50','drawInsufficient'].includes(state.status)) return;

  const col = sq.charCodeAt(0) - 97;
  const row = parseInt(sq[1]) - 1;
  const piece = state.board[row][col];

  if (selected) {
    if (sq === selected) { selected = null; render(); return; }

    const validTos = state.legalMoves[selected] || [];
    const matching = validTos.filter(to => to.slice(0, 2) === sq);

    if (matching.length > 0) {
      // Check if promotion
      if (matching.some(to => to.length > 2)) {
        pendingPromo = { from: selected, to: sq };
        showPromotion(state.activeColor);
        selected = null;
        return;
      }
      executeMove(selected, sq, '');
      return;
    }

    // Switch selection to another own piece
    if (piece?.color === state.activeColor) { selected = sq; render(); return; }
    selected = null; render(); return;
  }

  if (piece?.color === state.activeColor) { selected = sq; render(); }
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

    const mode = document.querySelector('input[name="mode"]:checked').value;
    if (!isGameOver() && shouldAiMove(mode)) {
      setTimeout(triggerAI, 250);
    }
  } catch (e) { console.error(e); }
}

function isGameOver() {
  return ['checkmate','stalemate','draw50','drawInsufficient'].includes(state?.status);
}

function shouldAiMove(mode) {
  if (mode === '2p') return false;
  return (mode === 'white' && state.activeColor === 'black') ||
         (mode === 'black' && state.activeColor === 'white');
}

async function triggerAI() {
  if (!state || isGameOver()) return;
  const cap = s => s.charAt(0).toUpperCase() + s.slice(1);
  turnIndEl.innerHTML = `${cap(state.activeColor)} thinking<span class="thinking"></span>`;
  try {
    const res = await fetch('/api/ai', { method: 'POST' });
    if (!res.ok) return;
    state = await res.json();
    render();
  } catch (e) { console.error(e); }
}

// ── Promotion dialog ──────────────────────────────────────────────────────────
function showPromotion(color) {
  promoChoices.innerHTML = '';
  [['queen','q'],['rook','r'],['bishop','b'],['knight','n']].forEach(([type, code]) => {
    const btn = document.createElement('div');
    btn.className = 'promo-btn';
    btn.appendChild(pieceImg(color, type));
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
  state = await (await fetch('/api/new', { method: 'POST' })).json();
  render();
  const mode = document.querySelector('input[name="mode"]:checked').value;
  if (mode === 'black') setTimeout(triggerAI, 300);
});

btnUndo.addEventListener('click', async () => {
  selected = null;
  const mode = document.querySelector('input[name="mode"]:checked').value;
  // Undo two moves in AI mode (player + AI response)
  state = await (await fetch('/api/undo', { method: 'POST' })).json();
  if (mode !== '2p') {
    state = await (await fetch('/api/undo', { method: 'POST' })).json();
  }
  render();
});

document.querySelectorAll('input[name="mode"]').forEach(radio => {
  radio.addEventListener('change', () => {
    flipped = radio.value === 'black';
    render();
  });
});

// ── Init ──────────────────────────────────────────────────────────────────────
(async () => {
  state = await (await fetch('/api/state')).json();
  render();
})();
