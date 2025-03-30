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
		// Get the form data from the request
		const formData = await request.formData();
		const file = formData.get('file') as File;

		if (!file) {
			return NextResponse.json({ error: 'No file uploaded' }, { status: 400 });
		}

		// Define the upload directory
		const uploadDir = path.join(process.cwd(), 'public/uploads');
		if (!fs.existsSync(uploadDir)) {
			fs.mkdirSync(uploadDir, { recursive: true }); // Create folder if not exists
		}

		// Save the uploaded file to /public/uploads
		const filePath = path.join(uploadDir, file.name);
		const fileBuffer = Buffer.from(await file.arrayBuffer());
		fs.writeFileSync(filePath, fileBuffer);

		console.log(`File saved to ${filePath}`);

		// Return the URL to access the uploaded file
		const fileUrl = `/uploads/${file.name}`;
		return NextResponse.json({ success: true, url: fileUrl });
	} catch (error) {
		console.error('Upload error:', error);
		return NextResponse.json({ error: 'Failed to upload file' }, { status: 500 });
	}
}

export async function GET() {
	try {
		// Define the upload directory
		const uploadDir = path.join(process.cwd(), 'public/uploads');

		// Check if the directory exists
		if (!fs.existsSync(uploadDir)) {
			return NextResponse.json({ files: [] }); // Return an empty array if no uploads
		}

		// Read all files in the upload directory
		const files = fs.readdirSync(uploadDir);

		// Create an array of objects with file paths
		const filePaths = files.map((file) => ({ path: `/uploads/${file}` }));

		return NextResponse.json({ files: filePaths });
	} catch (error) {
		console.error('Error reading uploads folder:', error);
		return NextResponse.json({ error: 'Failed to retrieve files' }, { status: 500 });
	}
}