//
// Filename: page.tsx
// Route: /
// Created: 3/29/25
//

"use client";

import { Bubble } from "pixel-retroui";

import '@/lib/pixel-retroui-setup.js';

export default function Home() {
	return (
		<div>
			<main>
				<div className="justify-center items-center my-[300px] flex flex-col">
					<Bubble
						className="flex justify-center items-center w-1/4"
						direction="left"
						onClick={ () => {
							console.log("Clicked");
						}}
					>
						Upload your image
					</Bubble>
				</div>
			</main>
		</div>
	);
}
