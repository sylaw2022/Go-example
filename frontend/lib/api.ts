import { User } from '@/types/user';

const API_BASE = process.env.API_URL || 'http://localhost:9090';

export async function getUsers(): Promise<User[]> {
    const res = await fetch(`${API_BASE}/api/v1/users`, {
        // Revalidate data every 60 seconds (ISR)
        next: { revalidate: 60 },
    });

    if (!res.ok) {
        throw new Error(`Failed to fetch users: ${res.status} ${res.statusText}`);
    }

    return res.json();
}
