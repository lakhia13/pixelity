//
// Filename: auth.ts
// Description: NextAuth options
// Created: 3/29/25
//

import { AuthOptions } from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";
import GoogleProvider from "next-auth/providers/google";

export const authOptions = {
	providers: [
		GoogleProvider({
			clientId: process.env.GOOGLE_CLIENT_ID!,
			clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
			authorization: {
				params: {
					prompt: "consent",
					access_type: "offline",
					response_type: "code",
					scope: "openid email profile"
				}
			},
			checks: ['none']
		}),

		CredentialsProvider({
			name: "Credentials",
			credentials: {
				username: { label: "Username", type: "text" },
				password: { label: "Password", type: "password" },
			},

			async authorize(credentials) {
				if (!credentials) {
					return null;
				}

				return {
					id: "",
					username: credentials.username,
					password: credentials.password,
				};
			},
		}),
	],
	
	// Write custom callbacks to deal with the custom session and user data
	callbacks: {
		async redirect({ baseUrl }: { baseUrl: string }): Promise<string> {
			return `${baseUrl}/view`;
		},

		async session({ session, token }: { session: any, token: any }): Promise<any> {
			if (token) {
				session.user.id 		= token.id;
				session.user.username 	= token.username;
				session.user.password 	= token.password;
			}
			
			return session;
		},

		async jwt({ token, user }: { token: any, user?: any }): Promise<any> {
			if (user) {
				token.id 		= user.id;
				token.username 	= user.username;
			}
			
			return token;
		},
	},

	session: {
		strategy: 	"jwt",
		maxAge: 	5184000, 	// Make users reauthenticate every 60 days
		updateAge: 	43200, 		// Update session every 12 hours
	},

	jwt: {
		maxAge: 	5184000, 	// Refresh the JWT session every 60 days
	},

	secret: process.env.NEXTAUTH_SECRET
} satisfies AuthOptions;