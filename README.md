A TCP server/client to track a retro board. Very much a WIP just learning about low level TCP operations.

Custom packet protocol: 1 byte version, 1 byte type, 4 byte payload size, payload. 
Custom Marshal/Unmarshal for each message type.

Client handled using Charm's Bubble Tea.
