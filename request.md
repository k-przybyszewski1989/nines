Let's plan together the app which is a game to play on a chess sheet.

## Technical requirements
1. Vue.js
2. Backend GoLang
3. Database MariaDB
4. Docker
5. CI/CD (GitHub Actions)

## Game rules
1. Board size 8x8 (X: A-H, Y: 1-8)
2. 9 pawns per side (Black and White)
3. Staring position: 
    - Home #1 -> Whites (Position: H1:F3)
    - Home #2 -> Blaks (Square's diagonal: A8:C6)
4. 18 pawns are always on the board
5. Pawn can never move on diagonal field to field, only forward and left
6. Pawn can always hop over (jump) other pawns no matter the color but only if the hop-by pawn is close enough, and a hop direction is forward or left.
7. Pawns can do multiple hops/jumps at once but never through more than 1 other pawn
8. 1 move per player
9. Backward moves are not allowed
10. Non-negotiable - 1 pawn per 1 field
11. Whites always start

### Fields selection
Whenever I say <LETTER><NUMBER>:<LETTER><NUMBER> this meeans I want you to think of big square containing all of the fields within that range.
Examples:
- H1:F3 (9 fields)
- H1:F2 (6 fields)

### Backward move
Every move backward and right is a backward move. Not allowed.

## Game goal
The first player to move 9 pawns to the opposite side of the board, Home #1 to Home #2 or Home #2 to Home #1 wins.

## User interface
1. User can select the game mode (Multiplayer or Singleplayer) at the very beginning
    - Singleplayer mode: Player(Human) vs Player(Computer)
    - Multiplayer mode: "Create a room" or join the room
2. Every player can see his name, counter of already made moves
3. Pawn selected to be moved by the player should be highlighted and then the player should be allowed to select the destination square or squares in case of multiple jumps. Once there is no other possible hop releasing last selected square does the move.

## Game modes
1. Multiplayer
2. Singleplayer

Use AskUserQuestionTool to ask me clarifying questions. And save the plan into plan.md before execution.
