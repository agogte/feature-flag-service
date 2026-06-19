const FLAG_API = "http://localhost:3000";

// Simulated user session (in a real app this comes from auth)
const USER_ID = "user_" + Math.floor(Math.random() * 1000);
console.log("userId:", USER_ID);

async function fetchFlag(flagKey, context) {
  try {
    const params = new URLSearchParams({ userId: context.userId });
    const res = await fetch(`${FLAG_API}/flags/${flagKey}/evaluate?${params}`);
    if (!res.ok) return false;
    const data = await res.json();
    return data.isEnabled ?? false;
  } catch {
    return false; //if api is unreachable, default to light mode
  }
}

async function init() {
  const darkModeEnabled = await fetchFlag("dark-mode", { userId: USER_ID });
  console.log("dark-mode enabled:", darkModeEnabled);

  if (darkModeEnabled) {
    document.body.classList.add("dark");
  }
}

init();

document.querySelectorAll(".lang-tabs").forEach((tabGroup) => {
  const group = tabGroup.dataset.group;
  const tabs = tabGroup.querySelectorAll(".lang-tab");
  const snippets = document.querySelectorAll(`.snippet[data-group="${group}"]`);

  tabs.forEach((tab) => {
    tab.addEventListener("click", () => {
      const lang = tab.dataset.lang;

      tabs.forEach((t) => t.classList.remove("active"));
      tab.classList.add("active");

      snippets.forEach((snippet) => {
        snippet.closest("pre").style.display =
          snippet.dataset.lang === lang ? "block" : "none";
      });
    });
  });
});
