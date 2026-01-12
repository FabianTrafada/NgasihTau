import { NextResponse } from 'next/server';

/**
 * API route to resend email verification
 * POST /api/auth/resend-verification
 * 
 * Body: { email: string }
 * 
 * NOTE: Backend requires authentication for /send-verification endpoint.
 * This means user must be logged in to resend verification.
 * For now, we'll try to call it but it may fail if user has no token.
 */
export async function POST(request: Request) {
    try {
        const body = await request.json();
        const { email } = body;

        if (!email) {
            return NextResponse.json(
                { error: { message: 'Email is required' } },
                { status: 400 }
            );
        }

        // Get backend API URL
        const apiUrl = process.env.NEXT_PUBLIC_API_URL?.endsWith('/')
            ? process.env.NEXT_PUBLIC_API_URL
            : `${process.env.NEXT_PUBLIC_API_URL}/`;

        // Try to get access token from request cookies/headers
        const accessToken = request.headers.get('Authorization')?.replace('Bearer ', '');

        // Forward request to backend with token if available
        const headers: HeadersInit = {
            'Content-Type': 'application/json',
        };

        if (accessToken) {
            headers['Authorization'] = `Bearer ${accessToken}`;
        }

        const response = await fetch(`${apiUrl}api/v1/auth/send-verification`, {
            method: 'POST',
            headers,
        });

        const data = await response.json();

        if (!response.ok) {
            // If unauthorized, return more helpful error
            if (response.status === 401) {
                return NextResponse.json(
                    {
                        error: {
                            message: 'Please log in again to resend verification email',
                            code: 'AUTHENTICATION_REQUIRED'
                        }
                    },
                    { status: 401 }
                );
            }
            return NextResponse.json(data, { status: response.status });
        }

        return NextResponse.json(data);
    } catch (error) {
        console.error('Resend verification error:', error);
        return NextResponse.json(
            { error: { message: 'Failed to resend verification email' } },
            { status: 500 }
        );
    }
}
