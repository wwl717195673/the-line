import { requestJSON } from "../lib/http";
import type { CreatePersonInput, PageData, Person, PersonQuery, UpdatePersonInput } from "../types/api";

export function listPersons(query: PersonQuery): Promise<PageData<Person>> {
  return requestJSON<PageData<Person>>("/api/persons", undefined, query);
}

export function createPerson(input: CreatePersonInput): Promise<Person> {
  return requestJSON<Person>("/api/persons", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function updatePerson(id: number, input: UpdatePersonInput): Promise<Person> {
  return requestJSON<Person>(`/api/persons/${id}`, {
    method: "PUT",
    body: JSON.stringify(input)
  });
}

export function disablePerson(id: number): Promise<Person> {
  return requestJSON<Person>(`/api/persons/${id}/disable`, {
    method: "POST",
    body: JSON.stringify({})
  });
}
