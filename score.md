# Plan: System punktacji ruchów w PlayerInfo

## Context
Zamiast licznika ruchów w komponencie PlayerInfo wyświetlamy sumę punktów za wykonane ruchy.
Punkty zależą od typu ruchu (krok prosty, przeskok prosty, przeskok ukośny) i pozycji pionków,
przez które gracz przeskakuje.

---

## Reguły punktacji

### Wartości bazowe kwadratów
```
col_val: A=1, B=1, C=2, D=3, E=3, F=2, G=1, H=1
row_val: 1=1, 2=1, 3=2, 4=3, 5=3, 6=2, 7=1, 8=1
baseValue(pos) = max(col_val, row_val)
```
Np. E4 = max(3,3)=3, G3 = max(1,2)=2.

### Typy ruchów
| Typ | Warunek | Wynik |
|-----|---------|-------|
| Krok zwykły | distance=1, brak przeskoku | 1 punkt |
| Przeskoki proste (tylko pionowo/poziomo) | D=0, S≥1 | sum(baseValue(midpoints)) |
| Przeskoki ukośne | D≥1, S=0 | (sum_ukośnych)^(D+1) — minimalne potęgowanie to **^2** |
| Kombinacja ukośnych + prostych | D≥1, S≥1 | (sum_ukośnych)^(D+1) + sum_prostych |

> **Ważne:** Potęga nigdy nie jest ^1. Minimalny wykładnik = 2 (dla D=1: ^2, D=2: ^3, D=3: ^4 …).

Gdzie:
- D = liczba przeskoków ukośnych w ruchu
- S = liczba przeskoków prostych (pionowych/poziomych) w ruchu
- `midpoint` = kwadrat pomiędzy startową a końcową pozycją każdego przeskoku

### Przykłady weryfikacji
- Prosty pionowy, 2 pionki (G4 val=3, G3 val=2): 3+2 = 5 ✓
- 1 ukośny przeskok nad G3 (val=2): 2^(1+1) = 4 ✓
- 3 ukośne przeskoki, wartości 1,2,3: (1+2+3)^(3+1) = 6^4 = 1296 ✓

---

## Pliki do modyfikacji

### 1. NOWY PLIK: `frontend/src/utils/score.ts`
Pura logika punktacji (bez zależności od Vue/Pinia):
```typescript
const COL_VALUES: Record<string, number> = { A:1, B:1, C:2, D:3, E:3, F:2, G:1, H:1 }
const ROW_VALUES: Record<number, number> = { 1:1, 2:1, 3:2, 4:3, 5:3, 6:2, 7:1, 8:1 }

export function baseValue(pos: string): number
export function midpoint(a: string, b: string): string        // zwraca np. "G3"
export function isSimpleStep(from: string, to: string): boolean  // distance=1
export function isDiagonalHop(from: string, to: string): boolean // row AND col change
export function calculateMoveScore(from: string, path: string[]): number
```

Algorytm `calculateMoveScore`:
1. Jeśli `path.length === 1` i `isSimpleStep(from, path[0])`: zwróć 1
2. Iteruj pary (positions[i], positions[i+1]) gdzie positions = [from, ...path]
3. Dla każdej pary wylicz midpoint → baseValue(mid)
4. Podziel na diagonal (obie osie zmieniają się o 2) vs straight (jedna oś zmienia się o 2)
5. Wynik = (diagonalSum)^(diagonalCount+1) + straightSum

### 2. `frontend/src/stores/player.ts`
- Zmiana: `moveCount: ref(0)` → `totalScore: ref(0)`
- Zmiana: `incrementMoves()` → `addScore(points: number)`
- Dodanie: `aiScore: ref(0)` + `addAiScore(points: number)`
- `reset()` zeruje oba pola

### 3. `frontend/src/views/GameView.vue`
Zmodyfikować `onMoveCommit` (linia 211) i `blackMoves` computed (linia 153):

```typescript
// Zastąpić blackMoves computed:
// const blackMoves = computed(() => playerStore.aiScore)

// W onMoveCommit:
async function onMoveCommit(from: string, path: string[]) {
  if (gameStore.state?.mode === 'multiplayer') {
    gameStore.submitMoveWS(from, path)
    playerStore.addScore(calculateMoveScore(from, path))
  } else {
    await gameStore.submitMove(gameId.value, from, path)
    playerStore.addScore(calculateMoveScore(from, path))   // biały
    const lm = gameStore.state?.last_move
    if (lm?.player === 'black') {
      playerStore.addAiScore(calculateMoveScore(lm.from, lm.path))
    }
  }
}
```

Propsy PlayerInfo:
- Czarny: `:move-count="playerStore.aiScore"` (zamiast `blackMoves`)
- Biały: `:move-count="playerStore.totalScore"` (zamiast `playerStore.moveCount`)

### 4. `frontend/src/components/PlayerInfo.vue`
- Prop: `moveCount: number` → `score: number`
- Wyświetlanie: `{{ score }} pts` (zamiast `{{ moveCount }} move(s)`)

---

## Kolejność implementacji
1. `score.ts` (logika niezależna, można testować izolowanie)
2. `player.ts` (store)
3. `GameView.vue` (integracja)
4. `PlayerInfo.vue` (widok)

## Weryfikacja
- Krok zwykły → 1 pkt
- Przeskok pionowy nad jednym pionkiem na G4 (val=3) → 3 pkt
- Przeskok ukośny nad G3 (val=2) → 4 pkt
- 3 ukośne nad wartościami 1,2,3 → 1296 pkt
- Kombinacja 1 ukośnego (val=2) + 1 prostego (val=3) → 2^2 + 3 = 7 pkt
- Po restarcie gry: score = 0 dla obu graczy
