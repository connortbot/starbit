# ![Starbit](./screenshots/08.png)
# starbit `v0.03.0`
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
Players earn 1 GES (General Energy Substance) per system owned. GES is used to create new Fleets and Ships.

**Fleets**: Composed of several Ships. A starting Fleet begins with 1 Destroyer.
- Can only be created in Systems you control.
- Combat resolves automatically when Fleets from different players occupy the same system.
- During combat, each Fleet randomly selects an enemy to attack and deals damage each sol.
- A player wins by eliminating all opposing Fleets and controlling all systems.
- Stats are defined as an average of all the Ships.
- Construct new Ships in Fleets to modify stats.

| Ship | Cost | Health | Attack | Ex Attack | Evasion | Armor |
|-|-|-|-|-|-|-|
| Destroyer | 250 GES | 50 | 2 | 1 | 35% | 5% |
| Cruiser |  350 GES | 75 | 1 | 2 | 20% | 15% |
| Battleship | 800 GES | 200 | 5 | 2 | 10% | 30% |
| Dreadnought | 1500 GES | 600 | 3 | 5 | 5% | 40% |

The creation of a fleet costs 3000 base GES.

### Controls
- **Navigation**: Press `Shift+E` and use arrow keys to move around the galaxy.
- **Inspector**: Press `Shift+I` to open the inspector panel, and use arrow keys to scroll up and down.
- **Commands**: Press `Shift+C` to access the command line, where you can enter:
  - `fc <system>` - Create a new Fleet in the specified system
  - `fm <fleet id> <to system id>` - Move a Fleet from one system to another
  - `fu <fleet_id> <de|cr|ba|dr>` - Builds a ship and adds it to a fleet
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
   curl -L https://github.com/connortbot/starbit/releases/download/v0.03/starbit-server-linux -o ~/starbit/starbit-server
   
   # Or alternatively with wget:
   # wget https://github.com/connortbot/starbit/releases/download/v0.03/starbit-server-linux -O ~/starbit/starbit-server

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

#### `v0.03`: Combat Update
Revamps the combat system so that each fleet is comprised of multiple Ships.

Adds stats *Evasion* and *Armor* to fleets.
- *Armor* reduces the damage received by a percentage.
- *Evasion* is a percentage chance that the damage received is dodged entirely.

Adds 4 types of ships: Destroyer, Cruiser, Battleship, Dreadnought.

| Ship | Cost | Health | Attack | Ex Attack | Evasion | Armor |
|-|-|-|-|-|-|-|
| Destroyer | 250 GES | 50 | 2 | 1 | 35% | 5% |
| Cruiser |  350 GES | 75 | 1 | 2 | 20% | 15% |
| Battleship | 800 GES | 200 | 5 | 2 | 10% | 30% |
| Dreadnought | 1500 GES | 600 | 3 | 5 | 5% | 40% |

The new `fu` command can be used to add ships to a fleet.

#### Future Updates
- Automatically grant a player the win when they are the only remaining one with owned systems.
- Build Supply System, requiring *Convoys* scaling with GES/tick consumption, and supply penalties.
- Combat Bonuses (outnumbering, ownership of system, etc.)
- Restrict movement in the grid? (e.g choke points)
- `sel <system_id>` which selects a group of fleets
- `spl <t/b>` which selects `(t)top` half or `(b)bottom` half of current selection
- `sm <system_id>` which moves selections to the system.
- Check client and server versions when joining a server. (Reject connections with an error message)
- Commands to modify fleet compositions (transfer ships, sell ships, in bulk)

## Contributing
Contributions are welcome! Please open an issue or submit a pull request.
Contact me at (loiconnor8@gmail.com).
