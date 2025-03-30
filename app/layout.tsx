import type { Metadata } from "next";
import Minecraft from "next/font/local";

import NavBar from "@/components/navbar";

import "./globals.css";
import "@/lib/pixel-retroui-setup.js";

const minecraftFont = Minecraft({ src: "../public/fonts/Minecraft.otf" });

export const metadata: Metadata = {
	title: "Pixelity",
	description: "Self-hosted solution to all your photo storage needs; all on one phone.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode}>) {
	return (
		<html lang="en">
			<body className={`${minecraftFont.className} antialiased`}>
				<NavBar/>
				{children}
			</body>
		</html>
	);
}
