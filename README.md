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

See the [docs](https://monarch.gitbook.io/monarch) to learn more about Monarch and integrating your own projects with the framework.

## Installing Monarch

[Follow the installation instructions here.](https://monarch.gitbook.io/monarch/installation)

## Empress
Empress is the very first integration developed alongside Monarch as a proof of concept.
The techniques used to develop the implant and builder service should be viewed as best practice, along with recommendations and examples provided in the [documentation](https://monarch.gitbook.io/monarch/integration).

[Find Empress here.](https://github.com/pygrum/Empress)


## Issues
If you encounter issues of any sort, please raise a new issue in the 
[issues page](https://github.com/pygrum/monarch/issues), especially as this project is in its early stages of development.
I'll do my best to response and resolve the issue on time.

## Contributing
Feel free to contact me about wanting to contribute on the `#golang` channel on the BloodHoundGang slack (@Pygrum).

## Disclaimer
This Command and Control (C2) framework is intended for authorized and lawful use only. 
Any unauthorized or illegal activities facilitated by this software are strictly prohibited. 
The developers are not liable for any misuse or illegal actions performed with this framework.
Users must comply with all applicable laws and ethical standards when using this software. 
The developers disclaim responsibility for any damages or legal consequences resulting from its misuse.
By using this software, you agree to use it responsibly and strictly for lawful purposes.

## Credits

This project was heavily inspired by the following projects:
- [Mythic](https://github.com/its-a-feature/Mythic): @its-a-feature - Inspiration for Docker container usage
- [Sliver](https://github.com/BishopFox/sliver): @moloch-- - Awesome CLI and RPC implementations

Here are some cool packages I tried out:
- [Console](https://github.com/reeflective/console): @maxlandon - Great CLI
- [Grumble](https://github.com/desertbit/grumble): @desertbit - Another great CLI

Go and check them out!
