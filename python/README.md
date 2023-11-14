# Monarch
## Integration package

This package is used to create services that help monarch operate.
Each project typically comes with two services:
1. Builder service - responsible for building agents, and providing the C2 with possible build parameters
2. Translator service - responsible for encoding and decoding messages transmitted between the agent and C2

Thanks to the already-implemented classes, integration is as nearly as simple as writing a build
and translation routine.

<!--TODO: Write up documentation for both builder and translator packages>