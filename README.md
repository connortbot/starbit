# starbit
A lightweight, tick-based space strategy game in your terminal over gRPC.

# Case Study
This explains the process of `starbit`, implementation along the way. (kinda a Case Study, not really)

### Initial Goals:
- Create a space strategy game in <5 days.
- Implement stats-heavy systems in battling, supply, resources.
- Real time, complete games in 10 minutes or less.
- Lightweight, in the terminal
- Learn `gRPC`!
- (Optional) Uses keystrokes to control.

## Game Explanation
`starbit` takes 2-4 players per game. 
Each Player controls an respective Empire in a tiny, randomly-generated Galaxy. The Galaxy is made up of a grid of Systems. 

A Player wins by getting control over the entire Galaxy.

### Gameplay
Players begin at roughly 4 opposite ends of the Galaxy, each starting with 1 System, 1 Factory, and 1 Scout.
The game is updated in 5 second ticks. Players can perform as many actions as they want between each tick, and they all go simultaneously into effect in the next tick.

All actions by players can be divided into:
- Gaining resources
- Creating units
- Battling for control over Systems

The actions are performed via typing commands into the terminal.

#### Combat
Players move Fleets into Systems, which fight other enemy Fleets. During each tick, the Fleets will deal and take damage.
Once there is only one Player's Fleet remaining, that Player owns the Sector. 
Owning a Sector yields a 10% bonus to ALL stats.

Every Fleet has:
- Attack
- Ex Attack (high damage)
- Armor (weighted average % protection against Ex Attack)
- Evasion (weighted average % protection against Attack)
- Health (death upon 0, add more ships to increase)
Fleets are comprised of Ships, which can be added to Fleets thereby shifting the stats.

**Tick Resolution:**
1. Each Fleet rolls Evasion
2. Each Fleet chooses a random enemy Fleet and deals Attack. Mitigated by their Evasion.
3. Each Fleet then deals Ex Attack to the same Fleet. Mitigated by their Armor.
4. Every Fleet with Health <= 0, destroy.

Fleets receive a 50% penalty if they are not properly supplied. This leads us to...

#### Resources & Supply
There is only one currency in this game: General Energy Substance (GES).
Players build Factories, which generate GES/tick. GES is used to create Fleets, and all Fleets require spending of GES per tick.

Players must create Convoys. Each Convoy is able to 'supply' an amount of GES. 
**Example:** Jonathan spends around X GES/tick on Fleets. Each Convoy supplies around Y GES/tick. If Y < X, the Fleets get a supply penalty.
> Its important to note that Fixed costs (building Ships for Fleets, and Factories) don't contribute to Variable Costs (maintenance GES/tick)

Players can build an infinite amount of Factories. They are not located in Systems.


## Beginning
Started by creating a `gRPC` simple game service, where a server can send ticks every 5 seconds and clients can freely send messages.
Then added a lobby system, players can join and differentiated by username.

```
╭──── Starbit ────╮
│ Players: 1/2    │
├─────────────────┤
│ ★ bruh        
╰─────────────────╯

Press Ctrl+C to quit
```

For a very first, working version, I greatly simplified the core gameplay.
- No Supply system
- No GES or Factories
- No depth to Fleets, all same stats.

The biggest concern was the user experience. Originally, a Player would tens to hundreds of Ships individually managed. I wanted the gameplay to be keystroke heavy, but how could a player efficiently control over, say, 50 ships? 
A better way of designing this would be to lessen the amount of units that need to be controlled without sacrificing strategic depth in the stats.

> Instead of 50 ships, think of them as 3-5 Fleets! This is not only more realistic but easier to control.

### Basic UI
Added the ability to select cells and view a basic UI:
```
╭──── Starbit ─────╮
│ Players: 2/2     │
├──────────────────┤
│   connor1        │
│ ★ stella         │
├──────────────────┤
│      Ready       │
╰──────────────────╯

Galaxy Map:
■ □ □ □ □ □ □ □ □ □ 
□ □ □ □ □ □ □ □ □ □ 
□ □ □ □ □ □ □ □ □ □ 
□ □ □ □ □ □ □ □ □ □ 
□ □ □ □ □ □ □ □ □ □ 
□ □ □ □ □ □ □ □ □ □ 
□ □ □ □ □ □ □ □ □ □ 
□ □ □ □ □ □ □ □ □ □ 
□ □ □ □ □ □ □ □ □ □ 
□ □ □ □ □ □ □ □ □ ■ 

╭─ Command ────────────────────────────────────────────────╮
│ i love connor                                            │
╰──────────────────────────────────────────────────────────╯

Press Ctrl+C to quit
```
> courtesy of my girlfriend, also this is missing a lot of colours.
The next step is, since the cells is limiting in terms of information, to have anothe window to the right of the galaxy showing info of the selected cell.

This also means more UI custom functions to:
- wrap given string content in a box (splitting by lines, adding)
- placing boxes side by side

## Fleets
We initialize a basic Fleet type and some functions to spawn them in:
```protobuf
message Fleet {
  string owner = 1;
  int32 attack = 2;
  int32 health = 3;
}
```
And upon game start, we place 1 Fleet in the starting positions accordingly.

## User Controls
After some experimenting, I landed on having players:
- VIEW information with the Inspector
- CHOOSE information with the Galaxy Map, choosing the System
- ACTION with the Command line.

After basically creating a mini UI library from scratch to make the boxes and some keybinds to switch between modes, 

```
╭──── Starbit ─────╮
│ Players: 2/2     │
├──────────────────┤
│ ★ connortbot     │
│   pranavbedi     │
├──────────────────┤
│      Ready       │
╰──────────────────╯

╭─────── Galaxy ───────╮  ╭─────────────────────── Inspector ────────────────────────╮
│ ■ □ □ □ □ □ □ □ □ □  │  │ ╭────────╮ ╭──────────────────╮ ╭───────────────────╮    │
│ □ □ □ □ □ □ □ □ □ □  │  │ │ ID: 0  │ │ Location: 0, 0   │ │ Owner: connortbot │    │
│ □ □ □ □ □ □ □ □ □ □  │  │ ╰────────╯ ╰──────────────────╯ ╰───────────────────╯    │
│ □ □ □ □ □ □ □ □ □ □  │  │                                                          │
│ □ □ □ □ □ □ □ □ □ □  │  │ ╭──────────────────── Fleet ─────────────────────╮       │
│ □ □ □ □ □ □ □ □ □ □  │  │ │ HP: ██████████████████████████████████████ 100 │       │
│ □ □ □ □ □ □ □ □ □ □  │  │ │                                                │       │
│ □ □ □ □ □ □ □ □ □ □  │  │ │ Owner: connortbot    Attack: 10                │       │
│ □ □ □ □ □ □ □ □ □ □  │  │ ╰────────────────────────────────────────────────╯       │
│ □ □ □ □ □ □ □ □ □ ■  │  │                                                          │
╰──────────────────────╯  ╰──────────────────────────────────────────────────────────╯
╭─ Command ────────────────────────────────────────────────╮
│ >                                                        │
╰──────────────────────────────────────────────────────────╯

╭──────────────╮  ╭──────────────────╮  ╭──────────────────╮  ╭──────────────╮
│ Cmd: Shift+C │  │ Inspect: Shift+I │  │ Explore: Shift+E │  │ Quit: Ctrl+C │
╰──────────────╯  ╰──────────────────╯  ╰──────────────────╯  ╰──────────────╯

   Mode: Command
```
> again, no colours
