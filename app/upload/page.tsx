"use client";

import { Button } from "pixel-retroui";
import { FaFileUpload } from "react-icons/fa";
import { JSX } from "react";
import { useRouter } from "next/navigation";

export default function Upload(): JSX.Element {
	const router = useRouter();
	
	return (
		<div className="flex justify-center items-center mt-[100px]">
			<main className="flex justify-center items-center">
				<div className="flex flex-col items-center">
					<h1 className="text-2xl">Upload Files</h1>

					<form
						id="upload"
						className="pt-8"
						method="post"
						encType="multipart/form-data"
						onSubmit={async (e) => {
							e.preventDefault();
							const formData = new FormData(e.currentTarget);
							const file = formData.get("file");

							if (file && file instanceof File) {
								try {
									const response = await fetch("/api/upload", {
										method: "POST",
										body: formData,
									});

									if (response.ok) {
										console.log("File uploaded successfully");
										router.replace("/upload");
									} else {
										console.error("Failed to upload file");
									}
								} catch (error) {
									console.error("Error uploading file:", error);
								}
							} else {
								console.error("No file selected");
							}
						}}
					>
						<div>
							<label className="cursor-pointer" htmlFor="file">
								<div className="rounded-lg w-full bg-red-400 p-4">
									<span className="flex items-center justify-center gap-x-4">
										Upload
										<FaFileUpload size={30} />
									</span>
								</div>
							</label>

							<input type="file" id="file" name="file" />
						</div>
						<div>
							<Button type="submit" form="upload">
								Submit
							</Button>
						</div>
					</form>
				</div>
			</main>
		</div>
	)
}