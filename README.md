# ![Starbit](./screenshots/08.png)
# starbit `v0.02.0`
A lightweight, space RTS game played in the terminal using QUIC and gRPC.
Written fully in Go.

## 📚 Table of Contents
- [Game Overview](#game-overview)
- [How to Play](#how-to-play)
- [Controls](#controls)
- [Deployment](#deployment)
- [Roadmap](#roadmap)
- [Contributing](#contributing)

## Game Overview
`starbit` is a space real-time strategy game where 2-4 players compete to conquer the galaxy. Each player controls an Empire and aims to gain control over the entire galaxy by building and moving fleets to battle opponents.

### How to Play
A game takes place across multiple `sols`, a unit of time. Think of it like a tick, where every ~500ms, everything updates (e.g battles, movement, etc.)

Players begin with a starting System and a *Fleet*.

**GES**:
Players earn 1 GES (General Energy Substance) per system owned. GES is used to create new Fleets, which cost 2000 GES each.

**Fleets**: All Fleets start with 100 health and 1 attack.
- Can only be created in Systems you control.
- Combat resolves automatically when Fleets from different players occupy the same system.
- During combat, each ship randomly selects an enemy to attack and deals damage each sol.
- A player wins by eliminating all opposing Fleets and controlling all systems.

### Controls
- **Navigation**: Press `Shift+E` and use arrow keys to move around the galaxy.
- **Inspector**: Press `Shift+I` to open the inspector panel, and use arrow keys to scroll up and down.
- **Commands**: Press `Shift+C` to access the command line, where you can enter:
  - `fc <system>` - Create a new Fleet in the specified system
  - `fm <fleet id> <to system id>` - Move a Fleet from one system to another
- **Fleets**: Press `Shift+F` to select the fleets list, and arrow keys to scroll up and down.

## Deployment
You'll need:
- AWS Account (Free Tier is fine)
- Terraform (or you can do it manually)

1. **Create a key-pair** in AWS Ohio (`us-east-2`). Download the `.pem` file and place it in `~/.ssh/starbit.pem`.

2. **Set up the infrastructure**:
   ```shell
   cd infrastructure
   terraform init
   terraform apply -var="key_name=starbit"
   ```

3. **Connect to the server**:
   ```shell
   ssh -i ~/.ssh/starbit.pem ubuntu@<IP>
   ```

### Option 1: Using Pre-built Server Executable (Recommended)

4. **Download and run the server directly** (recommended):
   ```shell
   # Create a directory for the server
   mkdir -p ~/starbit

   echo "net.core.rmem_max=8388608
   net.core.wmem_max=8388608" | sudo tee -a /etc/sysctl.conf
   sudo sysctl -p
   
   # Download the executable directly to the server
   curl -L https://github.com/connortbot/starbit/releases/download/v0.02/starbit-server-linux -o ~/starbit/starbit-server
   
   # Or alternatively with wget:
   # wget https://github.com/connortbot/starbit/releases/download/v0.02/starbit-server-linux -O ~/starbit/starbit-server

   # To run with a specific amoutn of players (max 4)
   ~/starbit/starbit-server -maxPlayers=2
   
   # Make it executable
   chmod +x ~/starbit/starbit-server
   
   # Run the server
   ~/starbit/starbit-server
   ```
5. **Destroy the infrastructure** when done:
   ```shell
   terraform destroy -var="key_name=starbit"
   ```

### Option 2: Building from Source

4. **Clone the repository and set up the server**:
   ```shell
   git clone https://github.com/connortbot/starbit.git
   cd starbit
   chmod +x setup_server.sh
   ./setup_server.sh
   source ~/.bashrc
   chmod +x run_server.sh
   ./run_server.sh <num_players>
   ```

5. **Destroy the infrastructure** when done:
   ```shell
   terraform destroy -var="key_name=starbit"
   ```

## Roadmap

#### `v0.02` Patch Notes
Fleets:
- Movement Cooldown: 0 Sols -> 10 sols
- Cost: 500 GES -> 2000 GES

Misc Updates:
- Win Condition: `(galaxyHeight) * galaxyWidth)` -> `(galaxyHeight - 1) * galaxyWidth)`
- GES/sol: 1 -> 0
- GES/system: 0 -> 1
- Added a list showing your fleets on the right side, including current locations for easy `fm` commands!
- Added a first screen to notify players to make their terminal window large enough! Unfortunately, we can't set it via code.

Bug Fixes:
- Fixed bug where Inspector window would get too long because of the content of a previously long System info. Window now scrolls to top when switching Systems.
- Fixed bug where earlier joined players lobby list would not show players who joined until the game already started.
- Fixed bug where first few players to join don't colour in systems correctly for enemies.

**Update v0.02.01**
Change `fm x y z` to just `fm <id> <destination>` since we already track owned fleets with the window. e

#### Future Updates
- Automatically grant a player the win when they are the only remaining one with owned systems.
- Add Ex(plosive) Attack, Evasion, and Armor.
- Ships (Destroyer, Cruiser, Battleship, Dreadnought) and Fleet composition of Ships.
- Build Supply System, requiring *Convoys* scaling with GES/tick consumption, and supply penalties.
- Combat Bonuses (outnumbering, ownership of system, etc.)
- Restrict movement in the grid? (e.g choke points)
- `sel <system_id>` which selects a group of fleets
- `spl <t/b>` which selects `(t)top` half or `(b)bottom` half of current selection
- `sm <system_id>` which moves selections to the system.

## Contributing
Contributions are welcome! Please open an issue or submit a pull request.
Contact me at (loiconnor8@gmail.com).
