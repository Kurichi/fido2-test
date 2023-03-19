import base64url from "base64url";
import React, { useState } from "react";

const Home = () => {
  const [name, setName] = useState<string>("");

  const signUp = async () => {
    const res = await fetch(`http://localhost:8080/auth/register/${name}`, {
      method: "GET",
      credentials: "include",
    });

    const te = new TextEncoder();
    const options = await res.json(); // as CredentialCreationOptions;
    if (options.publicKey) {
      options.publicKey.challenge = base64url.toBuffer(options.publicKey.challenge);
      options.publicKey.user.id = base64url.toBuffer(options.publicKey.user.id);
    }


    const credential = await navigator.credentials.create(options) as PublicKeyCredential;
    const td = new TextDecoder();
    const publicKeyCredential = {
      id: credential.id,
      type: credential.type,
      rawId: base64url.encode(credential.rawId as Buffer),
      response: {
        attestationObject: base64url.encode((credential.response as AuthenticatorAttestationResponse).attestationObject as Buffer),
        clientDataJSON: base64url.encode((credential.response as AuthenticatorAttestationResponse).clientDataJSON as Buffer),
      },
    }
    // console.log(publicKeyCredential)
    // console.log(JSON.stringify(publicKeyCredential))

    const response = await fetch(`http://localhost:8080/auth/register/${name}`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(publicKeyCredential),
    });

    console.log(await response.json());


  }

  return (
    <div>
      <h1>Home</h1>
      <input type="text" value={name} onChange={(e) => setName(e.target.value)} />
      <button
        onClick={() => signUp()}
      >
        SignUp
      </button>
    </div>
  );
};

export default Home;
