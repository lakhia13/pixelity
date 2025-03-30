//
// Filename: page.tsx
// Route: /api/auth/[...nextauth]
// Created: 3/29/25
//

import NextAuth from "next-auth";
import { authOptions } from "@/lib/auth";

const handler = NextAuth(authOptions);

export { handler as GET, handler as POST };