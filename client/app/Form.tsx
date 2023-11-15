"use client";
import { Button } from "@recipes/button";
import { Input } from "@recipes/input";
import { Label } from "@recipes/label";
import { Stack } from "@recipes/stack";
import { Textarea } from "@recipes/textarea";
import { useEffect, useRef, useState } from "react";

export function Form({ action }: { action: any }) {
  let formRef = useRef<HTMLFormElement | null>(null);

  let [timeZone, setTimeZone] = useState<string>("");

  useEffect(() => {
    setTimeZone(new Intl.DateTimeFormat().resolvedOptions().timeZone);
  }, []);

  return (
    <form
      action={async (formData) => {
        await action(formData);
        formRef.current?.reset();
      }}
      ref={formRef}
    >
      <input className="hidden" name="timezone" id="timezone" value={timeZone} readOnly />
      <Stack gap={10}>
        <div className="grid w-full items-center gap-4">
          <Label htmlFor="body">Update</Label>
          <Textarea id="body" name="body" placeholder="What have you been up to?" required />
        </div>
        <div className="grid w-full items-center gap-4">
          <Label htmlFor="location">Location</Label>
          <Input type="text" id="location" name="location" placeholder="Location" />
        </div>
        {
          /* <div className="grid w-full items-center gap-4">
          <Label htmlFor="media">Media</Label>
          <Input id="media" name="media" multiple accept="image/*,video/*" type="file" />
        </div> */
        }
        <Button
          variant="outline"
          type="submit"
        >
          Submit
        </Button>
      </Stack>
    </form>
  );
}
