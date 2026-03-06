import { useState, useEffect } from 'react'
import api from './lib/api'

function App() {
    const [users, setUsers] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)

    useEffect(() => {
        const fetchUsers = async () => {
            try {
                const response = await api.get('/api/v1/users')
                setUsers(response.data)
            } catch (err) {
                setError('Failed to fetch data from the server. Is the backend running?')
            } finally {
                setLoading(false)
            }
        }
        fetchUsers()
    }, [])

    return (
        <div className="app-container">
            <header className="hero-section">
                <h1 className="hero-title">Industrial Standard Fullstack</h1>
                <p className="hero-subtitle">Go Backend &middot; React Frontend</p>
            </header>

            <main className="main-content">
                <section className="data-section glass-panel">
                    <h2>Registered Users</h2>

                    {loading && <div className="loader"></div>}

                    {error && <div className="error-message">{error}</div>}

                    {!loading && !error && (
                        <div className="user-grid">
                            {users.map(user => (
                                <div key={user.id} className="user-card">
                                    <div className="user-avatar">
                                        {user.name.charAt(0)}
                                    </div>
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
    )
}

export default App
