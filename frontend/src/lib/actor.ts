type Actor = {
  personId?: number;
  roleType?: string;
};

const STORAGE_KEY = "the_line_actor";

export function getActor(): Actor {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return {};
    }
    const parsed = JSON.parse(raw) as Actor;
    return {
      personId: parsed.personId && parsed.personId > 0 ? parsed.personId : undefined,
      roleType: parsed.roleType?.trim() || undefined
    };
  } catch {
    return {};
  }
}

export function setActor(next: Actor): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
}
