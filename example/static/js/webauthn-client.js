class WebauthnClient {
    constructor(loginUrl, registerUrl) {
        this.loginUrl = loginUrl;
        this.registerUrl = registerUrl;
    }

    async register(headers) {
        let response = await fetch(this.registerUrl, {
            method: 'GET',
            headers: headers
        });
        let data = await response.json();
        let responseHeaders = response.headers;
        headers["Webauthn-Session"] = responseHeaders.get("Webauthn-Session");
        headers["Content-Type"] = "application/json";
        data.publicKey.challenge = this.decode(data.publicKey.challenge);
        data.publicKey.user.id = this.decode(data.publicKey.user.id);
        let result = await navigator.credentials.create({
            publicKey: data.publicKey
        });

        let rawId = this.encode(result.rawId);
        let payLoad = JSON.stringify({
            id: result.id,
            rawId: rawId,
            type: result.type,
            response: {
                attestationObject: this.encode(result.response.attestationObject),
                clientDataJSON: this.encode(result.response.clientDataJSON)
            }
        });
        fetch(this.registerUrl, {
                method: 'POST',
                headers: headers,
                body: payLoad
            }
        ).then(success => {
            return success.ok;
        }, () => {
            // TODO: Proper handling?
            console.log("BONKERS");
            return false;
        });
    }

    async login(headers) {
        let response = await fetch(this.loginUrl,{
                method: 'GET',
                headers: headers
            });
        let data = await response.json();
        let responseHeaders = response.headers;
        headers["Webauthn-Session"] = responseHeaders.get("Webauthn-Session");
        headers["Content-Type"] = "application/json";

        data.challenge = this.decode(data.challenge);
        data.allowCredentials.forEach(listItem => {
            listItem.id = this.decode(listItem.id)
        }, this);
        let result = await navigator.credentials.get({
            publicKey: data
        });
        {
            let payLoad = JSON.stringify({
                id: result.id,
                rawId: this.encode(result.rawId),
                type: result.type,
                response: {
                    authenticatorData: this.encode(result.response.authenticatorData),
                    clientDataJSON: this.encode(result.response.clientDataJSON),
                    signature: this.encode(result.response.signature),
                    userHandle: this.encode(result.response.userHandle)
                }
            });
            fetch(this.loginUrl, {
                method: 'POST',
                headers: headers,
                body: payLoad
            }).then(success => {
                return success.ok;
            }, () => {
                // TODO: Proper handling?
                console.log("BONKERS");
                return false;
            });
        }
    }

    encode(value) {
        return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
            .replace(/\+/g, "-")
            .replace(/\//g, "_")
            .replace(/=/g, "");
    }

    decode(value) {
        return Uint8Array.from(atob(value), c => c.charCodeAt(0));
    }

}
