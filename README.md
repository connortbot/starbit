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
- Battling for control over sectors

The actions are performed via typing commands into the terminal.

#### Combat
Players move Ships into Systems, which fight other enemy Ships. During each tick, the Ships will deal and take damage.
Once there is only one Player's Ships remaining, that Player owns the Sector. 
Owning a Sector yields a 10% bonus to ALL stats.

Every Ship has:
- Attack
- Ex Attack (high damage)
- Armor (% protection against Ex Attack)
- Evasion (% chance to avoid Attack)
- Health (death upon 0)
- Recovery Rate

**Tick Resolution:**
1. Each Ship rolls Evasion
2. Each Ship chooses a random enemy Ship and deals Attack. Deals 0 if they evade.
3. Each Ship then deals Ex Attack to the same Ship. Mitigated by their Armor.
4. Every Ship with Health <= 0, destroy.

When not in battles, all Ships regain Health at the Recovery Rate if they own the System.
Ships receive a 50% penalty if they are not properly supplied. This leads us to...

#### Resources & Supply
There is only one currency in this game: General Energy Substance (GES).
Players build Factories, which generate GES/tick. GES is used to create Ships, and all Ships require spending of GES per tick.

Players must create Convoys. Each Convoy is able to 'supply' an amount of GES. 
**Example:** Jonathan spends around X GES/tick on Ships. Each Convoy supplies around Y GES/tick. If Y < X, the Ships get a supply penalty.
> Its important to note that Fixed costs (building Ships and Factories) don't contribute to Variable Costs (maintenance GES/tick)

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
- Only one type of Ship
- No Supply system
- No GES or Factories

The biggest concern was the user experience. I wanted the gameplay to be keystroke heavy, but how could a player efficiently control over, say, 50 ships? 
A better way of designing this would be to lessen the amount of units that need to be controlled withought sacrificing strategic depth in the stats.

Instead of 50 ships, think of them as 3-5 fleets. This is not only more realistic but easier to control.
