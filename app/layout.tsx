import type { Metadata } from "next";
import Minecraft from "next/font/local";

import "./globals.css";

const minecraftFont = Minecraft({ src: "../public/fonts/Minecraft.otf" });

export const metadata: Metadata = {
	title: "Pixility",
	description: "Self-hosted solution to all your photo storage needs; all on one phone.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode}>) {
	return (
		<html lang="en">
			<body className={`${minecraftFont.className} antialiased`}>
				{children}
			</body>
		</html>
	);
}
