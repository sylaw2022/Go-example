import { getUsers } from '@/lib/api';
import { User } from '@/types/user';

// This is a React Server Component — data is fetched on the server.
// No CORS issues, no client-side loading state needed.
export default async function HomePage() {
  let users: User[] = [];
  let error: string | null = null;

  try {
    users = await getUsers();
  } catch (err) {
    error = 'Failed to fetch data from the server. Is the backend running?';
  }

  return (
    <div className="app-container">
      <header className="hero-section">
        <h1 className="hero-title">Industrial Standard Fullstack</h1>
        <p className="hero-subtitle">Go Backend &middot; Next.js Frontend</p>
      </header>

      <main className="main-content">
        <section className="data-section glass-panel">
          <h2>Registered Users</h2>

          {error && <div className="error-message">{error}</div>}

          {!error && (
            <div className="user-grid">
              {users.map((user) => (
                <div key={user.id} className="user-card">
                  <div className="user-avatar">{user.name.charAt(0)}</div>
                  <div className="user-info">
                    <h3>{user.name}</h3>
                    <p>{user.email}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
