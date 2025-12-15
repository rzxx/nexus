import { useState } from "react";
import { Chat } from "./Chat";

function App() {
  const [username, setUsername] = useState("Guest");
  const [isChatting, setIsChatting] = useState(false);

  return (
    <div style={{ padding: 20, fontFamily: "sans-serif" }}>
      <h1>Nexus Hybrid Framework</h1>
      <p>
        Architecture: React (Vite) - Hono (Serverless) - Go Engine (Stateful)
      </p>

      {!isChatting ? (
        <div>
          <label>Choose Username: </label>
          <input
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            style={{ marginRight: 10 }}
          />
          <button onClick={() => setIsChatting(true)}>Join Chat</button>
        </div>
      ) : (
        <Chat username={username} />
      )}
    </div>
  );
}

export default App;
