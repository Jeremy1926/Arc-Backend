import crypto from "crypto"

const SECRET_KEY = Buffer.from([
    0x85, 0x3a, 0x1f, 0x7c, 0x2d, 0x9e, 0xab, 0x4b, 0x6f, 0x12, 0x34, 0x56, 0x78, 0x90, 0xaa, 0xbb,
    0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0xab, 0xcd,
    0xef
])

const obfInsertCount = 8

const deriveKey = salt =>
    crypto.createHash("sha256").update(Buffer.concat([SECRET_KEY, salt])).digest()

const base32Decode = input => {
    const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
    let bits = 0
    let value = 0
    const output = []
    for (const c of input.replace(/=+$/, "")) {
        const idx = alphabet.indexOf(c)
        if (idx === -1) continue
        value = (value << 5) | idx
        bits += 5
        if (bits >= 8) {
            output.push((value >>> (bits - 8)) & 0xff)
            bits -= 8
        }
    }
    return Buffer.from(output)
}

const obfPlan = origLen => {
    const lenBuf = Buffer.alloc(4)
    lenBuf.writeUInt32BE(origLen, 0)
    const sum = crypto.createHash("sha256").update(Buffer.concat([SECRET_KEY, lenBuf])).digest()
    const positions = new Set()
    for (let i = 0; i < obfInsertCount; i++) {
        if (origLen === 1) {
            positions.add(0)
            continue
        }
        let p = sum[i] % origLen
        while (positions.has(p)) p = (p + 1) % origLen
        positions.add(p)
    }
    return positions
}

const deobfuscate = obf => {
    if (obf.length < obfInsertCount) throw new Error("token too short")
    const origLen = obf.length - obfInsertCount
    if (origLen <= 0) throw new Error("invalid token length")
    const positions = obfPlan(origLen)
    let out = ""
    let j = 0
    for (let i = 0; i < origLen; i++) {
        if (positions.has(i)) {
            if (j >= obf.length) throw new Error("malformed token")
            j++
        }
        if (j >= obf.length) throw new Error("malformed token")
        out += obf[j++]
    }
    if (j !== obf.length) throw new Error("malformed token")
    return out
}

function decodeSession(token) {
    const layer3 = deobfuscate(token)
    const layer2 = base32Decode(layer3.toUpperCase()).toString()
    if (layer2.length % 2 !== 0) throw new Error("invalid hex length")

    const layer1 = Buffer.from(layer2, "hex").toString()
    const combined = Buffer.from(layer1, "base64url")

    if (combined.length < 16 + 12 + 64) throw new Error("token too short")

    const salt = combined.subarray(0, 16)
    const key = deriveKey(salt)

    const nonce = combined.subarray(16, 28)
    const body = combined.subarray(28)
    if (body.length < 64) throw new Error("missing signature")

    const ciphertextWithTag = body.subarray(0, body.length - 64)
    const signature = body.subarray(body.length - 64)

    const expected = crypto.createHmac("sha512", SECRET_KEY)
        .update(ciphertextWithTag)
        .digest()

    if (!crypto.timingSafeEqual(expected, signature))
        throw new Error("invalid signature")

    const tag = ciphertextWithTag.subarray(ciphertextWithTag.length - 16)
    const ciphertext = ciphertextWithTag.subarray(0, ciphertextWithTag.length - 16)

    const decipher = crypto.createDecipheriv("aes-256-gcm", key, nonce)
    decipher.setAuthTag(tag)

    const plaintext = Buffer.concat([
        decipher.update(ciphertext),
        decipher.final()
    ])

    return JSON.parse(plaintext.toString())
}

const token = process.argv[2]

if (!token) {
    console.error("usage: bun decode_sesh.js <token>")
    process.exit(1)
}

try {
    const session = decodeSession(token)
    console.log(JSON.stringify(session, null, 2))
} catch (e) {
    console.error("Error:", e.message)
    process.exit(1)
}