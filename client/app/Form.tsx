"use client";

import { Button } from "@recipes/button";
import { Input } from "@recipes/input";
import { Label } from "@recipes/label";
import { Stack } from "@recipes/stack";
import { Textarea } from "@recipes/textarea";
import { useFormStatus } from "react-dom";

export function Form() {
  let formStatus = useFormStatus();
  return (
    <Stack gap={10}>
      <div className="grid w-full items-center gap-4">
        <Label htmlFor="body">Update</Label>
        <Textarea id="body" placeholder="What have you been up to?" required />
      </div>
      <div className="grid w-full items-center gap-4">
        <Label htmlFor="location">Location</Label>
        <Input type="text" id="location" placeholder="Location" />
      </div>
      <Button
        variant={formStatus.pending ? "destructive" : "outline"}
        disabled={formStatus.pending}
        type="submit"
      >
        Submit
      </Button>
    </Stack>
  );
}
