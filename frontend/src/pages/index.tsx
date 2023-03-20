import base64url from "base64url";
import React, { useState } from "react";

const Home = () => {
  const [name, setName] = useState<string>("");

  const signUp = async () => {
    const res = await fetch(`http://localhost:8080/auth/register/${name}`, {
      method: "GET",
      credentials: "include",
    });

    const options = await res.json();
    if (options.publicKey) {
      options.publicKey.challenge = base64url.toBuffer(options.publicKey.challenge);
      options.publicKey.user.id = base64url.toBuffer(options.publicKey.user.id);
    }
    if (options.publicKey.excludeCredentials) {
      options.publicKey.excludeCredentials.forEach((listItem: any) => {
        listItem.id = base64url.toBuffer(listItem.id);
      });
    }

    const credential = await navigator.credentials.create(options) as PublicKeyCredential;
    const rawId = base64url.encode(credential.rawId as Buffer);
    const atteobj = base64url.encode((credential.response as AuthenticatorAttestationResponse).attestationObject as Buffer);
    const CDJ = base64url.encode((credential.response as AuthenticatorAttestationResponse).clientDataJSON as Buffer);

    const response = await fetch(`http://localhost:8080/auth/register/${name}`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        id: credential.id,
        type: credential.type,
        rawId: rawId,
        response: {
          attestationObject: atteobj,
          clientDataJSON: CDJ,
        },
      }),
    });

    console.log(await response.json());
  }

  const signIn = async () => {
    const res = await fetch(`http://localhost:8080/auth/login/${name}`, {
      method: "GET",
      credentials: "include",
    });

    const opt = await res.json(); // as CredentialRequestOptions;
    opt.publicKey.challenge = base64url.toBuffer(opt.publicKey.challenge);
    opt.publicKey.allowCredentials.forEach((listItem: any) => {
      listItem.id = base64url.toBuffer(listItem.id);
    })

    const credential = await navigator.credentials.get(opt) as PublicKeyCredential;
    const rawId = base64url.encode(credential.rawId as Buffer);
    const authData = base64url.encode((credential.response as AuthenticatorAssertionResponse).authenticatorData as Buffer);
    const CDJ = base64url.encode((credential.response as AuthenticatorAssertionResponse).clientDataJSON as Buffer);
    const sig = base64url.encode((credential.response as AuthenticatorAssertionResponse).signature as Buffer);
    const userHandle = base64url.encode((credential.response as AuthenticatorAssertionResponse).userHandle as Buffer);

    const response = await fetch(`http://localhost:8080/auth/login/${name}`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        id: credential.id,
        type: credential.type,
        rawId: rawId,
        response: {
          authenticatorData: authData,
          clientDataJSON: CDJ,
          signature: sig,
          userHandle: userHandle,
        },
      }),
    });

    console.log(await response.json());
  }

  return (
    <div>
      <h1>Home</h1>
      <input type="text" value={name} onChange={(e) => setName(e.target.value)} />
      <button onClick={signUp}>SignUp</button>
      <button onClick={signIn}>SignIn</button>
    </div>
  );
};

export default Home;
