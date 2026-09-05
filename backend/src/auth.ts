import { type Request, type Response } from 'express';

type TokenExchangeBody = {
  code: string;
  code_verifier: string;
  redirect_uri: string;
}

export const handlePostAuthToken = async (req: Request<{}, {}, TokenExchangeBody>, res: Response) => {
  const { code, code_verifier, redirect_uri } = req.body;

  if (!code || !code_verifier || !redirect_uri) {
    return res.status(400).json({ error: 'Missing required parameters.' });
  }
  if (!process.env.RC_TOKEN_URL || !process.env.RC_CLIENT_ID || !process.env.RC_CLIENT_SECRET) {
    return res.status(500).json({ error: 'Env not configured' });
  }

  try {
    const response = await fetch(process.env.RC_TOKEN_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        grant_type: 'authorization_code',
        client_id: process.env.RC_CLIENT_ID,
        client_secret: process.env.RC_CLIENT_SECRET,
        code,
        redirect_uri,
        code_verifier
      })
    });

    if (!response.ok) {
      console.error(`Auth server responded with status ${response.status}`);
      const responseText = await response.text();
      console.error(responseText);
      return res.status(response.status).json({ error: 'Auth server returned error.' });
    }

    const data = await response.json();
    return res.json(data);
  } catch (err) {
    console.error(`Token exchange error: ${err}`);
    return res.status(500).json({ error: 'Internal Server Error' });
  }
};
