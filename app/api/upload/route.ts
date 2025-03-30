//
// Filename: page.tsx
// Route: /api/upload
// Created: 3/30/25
//

import { NextResponse } from 'next/server';
import fs from 'fs';
import path from 'path';

export async function POST(request: Request) {
	try {
		const { filename, image } = await request.json(); // Get filename & base64 string

		if (!filename || !image) {
			return NextResponse.json({ error: 'Missing filename or image' }, { status: 400 });
		}

		// Decode base64 image
		const base64Data = image.replace(/^data:image\/\w+;base64,/, '');
		const buffer = Buffer.from(base64Data, 'base64');

		// Define the upload directory
		const uploadDir = path.join(process.cwd(), 'public/uploads');
		if (!fs.existsSync(uploadDir)) {
			fs.mkdirSync(uploadDir, { recursive: true }); // Create folder if not exists
		}

		// Save image to /public/uploads
		const filePath = path.join(uploadDir, filename);
		fs.writeFileSync(filePath, buffer);

		// Return the URL to access the image
		const imageUrl = `/uploads/${filename}`;
		return NextResponse.json({ success: true, url: imageUrl });
	} catch (error) {
		console.error('Upload error:', error);
		return NextResponse.json({ error: 'Failed to upload image' }, { status: 500 });
	}
}