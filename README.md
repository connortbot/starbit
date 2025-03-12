# starbit
A lightweight, tick-based space strategy game in your terminal over gRPC.

## Game
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