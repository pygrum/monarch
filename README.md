# Monarch
### The Adversary Emulation Framework

Monarch is a C2 Framework designed to give implant developers the convenience of integrating with an existing 
backend, so that more time can be spent creating cutting-edge features and enhancing overall efficiency. 
By seamlessly integrating with an established backend, Monarch enables developers to dedicate their efforts to 
refining and expanding the capabilities of their implants, ensuring a swift and efficient development process.

## How it works

Monarch leverages Docker containers to streamline the creation of agent builders, providing an isolated
environment for compilation. Using pre-installed RPC endpoints, Monarch abstracts out the effort of
building agents by providing an easy-to-use interface for managing build options and profiles.

Additionally, Monarch utilizes HTTP(s) endpoints to manage remote implants. 
These endpoints serve as a conduit, enabling efficient communication and control over distributed implants 
from a central hub. This approach empowers administrators to effectively oversee, direct, and interact with 
remote implants, facilitating smooth command execution and data retrieval.

## Installing Monarch
### System requirements
Monarch was primarily developed and tested on Ubuntu. Monarch will work on most Unix-based systems. 
For system requirements, see the requirements for Docker Desktop on Mac or Linux.

### Steps

1. Clone the repository
2. Run `bash scripts/install-monarch.sh`

Done! Monarch will be saved at `$HOME/.local/bin` for you to add to your `PATH`.

## Issues
If you encounter issues of any sort, please raise a new issue in the 
[issues page](https://github.com/pygrum/monarch/issues), especially as this project is in its early stages of development.
I'll do my best to response and resolve the issue on time.

## Contributing
Feel free to contact me about wanting to contribute on the `#golang` channel on the BloodHoundGang slack (@Pygrum).

## Inspirations

This project was heavily inspired by the following projects:
- [Mythic](https://github.com/its-a-feature/Mythic): @its-a-feature - Inspiration for Docker container usage
- [Sliver](https://github.com/BishopFox/sliver): @moloch-- - Awesome CLI

Go and check them out!