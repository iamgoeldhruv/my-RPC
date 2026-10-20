# My-RPC — Length-Prefix Framing Protocol

## Frame Format Design

### 1. Objective




The framing layer will define how My-RPC encodes individual messages into bytes and reconstructs complete messages when receiving data over TCP.

This document defines the initial frame format, payload size limits, and the expected behavior for edge cases. Implementation will follow in the next stage.

---

### 2. Design Decision

My-RPC will use **length-prefix framing**.

Each frame consists of:

```text
+---------------------------+----------------------+
| Length (4 bytes)          | Payload (N bytes)    |
+---------------------------+----------------------+
```

The first four bytes specify the length of the payload that immediately follows them.

The payload contains the message data. At this stage, it can be arbitrary bytes; it does not yet have a defined RPC-specific structure.

The framing layer is responsible only for determining where a message begins and ends. It does not interpret the payload.

---

### 3. Wire Format

The frame format is defined as follows:

| Field | Size | Description |
|---|---:|---|
| Length prefix | 4 bytes | Unsigned payload length in bytes |
| Payload | N bytes | Message content |
| Total frame size | 4 + N bytes | Prefix plus payload |

The length prefix represents the payload size only. It does not include the four bytes occupied by the prefix itself.

#### Byte order

The length prefix uses **big-endian byte order** (network byte order).

For example, a payload of 258 bytes has the following length prefix:

```text
00 00 01 02
```

The four bytes represent the unsigned integer 258.

Using a fixed-width prefix means the receiver always knows that it must read exactly four bytes to determine the expected payload length.

#### Example frame

Suppose the payload is the five-byte string `hello`.

```text
Length prefix:
00 00 00 05

Payload:
68 65 6C 6C 6F

Complete frame:
00 00 00 05 68 65 6C 6C 6F
```

The receiver reads the prefix, determines that the payload is five bytes long, and then reads exactly five payload bytes.

After reconstructing the complete frame, it can pass the payload to the next layer.

---

### 4. Maximum Payload Size

My-RPC will initially enforce a maximum payload size of **16 MiB (16 × 1024 × 1024 bytes)**.

```text
MAX_PAYLOAD_SIZE = 16,777,216 bytes
                  = 16 MiB
```

The four-byte unsigned length prefix can represent values from 0 through 4,294,967,295. However, the protocol intentionally imposes a much smaller application-level limit.

#### Why enforce a maximum?

An untrusted client could advertise an extremely large payload length. If the server attempted to allocate memory for that payload without validating the length, it could consume excessive memory or become vulnerable to resource-exhaustion attacks.

The maximum payload size therefore serves as a resource boundary for the framing layer.

The limit applies to the payload alone, excluding the four-byte prefix.

The implementation must validate the declared length before allocating memory for or reading the payload.

---

### 5. Edge Case Handling

#### Case 1: Zero-length payload

**Input frame:**

```text
00 00 00 00
```

The prefix declares a payload length of zero.

**Expected behavior:**

- The frame is structurally valid under the initial framing rules.
- The receiver returns an empty payload without attempting to read payload bytes.
- The connection remains open unless a higher-level protocol rule or application policy requires otherwise.

**Design decision:** Zero-length payloads are allowed at the framing layer.

This keeps framing independent of RPC semantics. A future RPC message format may choose to reject empty messages if they are not meaningful.

#### Case 2: Payload exceeds the maximum

Suppose a client sends a prefix declaring a payload length greater than 16 MiB.

**Expected behavior:**

1. Read the four-byte prefix.
2. Decode the declared payload length.
3. Compare it against `MAX_PAYLOAD_SIZE`.
4. Reject the frame immediately if the length exceeds the limit.
5. Do not allocate a buffer for the declared payload.
6. Close the connection and record an appropriate error.

**Design decision:** Oversized frames are rejected before payload allocation.

The receiver will not attempt to consume or discard the oversized payload. Closing the connection prevents the remaining bytes from being misinterpreted as a new frame.

#### Case 3: Partial reads from the network

TCP does not guarantee that a single write by the sender corresponds to a single read by the receiver.

For example, the sender might write:

```text
Write 1: 00 00
Write 2: 00 05 68
Write 3: 65 6C 6C 6F
```

The receiver may obtain these bytes in different-sized chunks.

It is also possible for the four-byte prefix itself to arrive across multiple reads.

**Expected behavior:**

- Read the complete four-byte prefix before decoding its length.
- If the prefix arrives in multiple chunks, continue reading until all four bytes have been obtained.
- Validate the decoded payload length.
- Read the complete payload, even if it arrives in multiple chunks.
- Return the payload only after all expected bytes have been received.
- Treat an unexpected EOF or read error before completion as an incomplete frame.



---

### 6. Frame Decoding Algorithm

The receiver will process each frame in the following order:

```text
Start
  |
  v
Read exactly 4 bytes
  |
  v
Decode unsigned big-endian length
  |
  v
Is length <= 16 MiB?
  |
  +---- No ----> Reject frame and close connection
  |
 Yes
  |
  v
Is length zero?
  |
  +---- Yes ---> Return empty payload
  |
  No
  |
  v
Allocate payload buffer
  |
  v
Read exactly the declared payload length
  |
  +---- Error --> Return error; do not return a partial payload
  |
  v
Return complete payload
```

The actual implementation must distinguish clean connection closure between frames from unexpected EOF in the middle of a frame.

If the connection closes before any byte of a new prefix is received, the decoder can report normal end-of-stream. If the prefix or payload is only partially received, the decoder must report an incomplete frame.

---

### 7. Error Handling Policy

The framing layer will use explicit errors to describe malformed or incomplete frames.

| Condition | Expected result | Connection policy |
|---|---|---|
| Valid non-empty payload | Return complete payload | Keep open |
| Zero-length payload | Return empty payload | Keep open |
| Length exceeds 16 MiB | Return oversized-frame error | Close |
| EOF before any prefix byte | Report clean end-of-stream | Close normally |
| EOF during the prefix | Report incomplete frame | Close |
| EOF during the payload | Report incomplete frame | Close |
| Read timeout during framing | Return timeout/read error | Close |
| Other network read error | Return the underlying error | Close |

The decoder must not return a partially reconstructed payload as a successful message.

For malformed or incomplete frames, the connection handler should terminate processing and allow the existing transport layer to close the connection.

---

