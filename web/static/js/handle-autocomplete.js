/**
 * Handle autocomplete for AT Protocol login
 * Provides typeahead search for Bluesky handles
 */
(function () {
  const input = document.getElementById("handle");
  const results = document.getElementById("autocomplete-results");

  // Exit early if elements don't exist (user might be authenticated)
  if (!input || !results) return;

  let debounceTimeout;
  let abortController;

  function debounce(func, wait) {
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(debounceTimeout);
        func(...args);
      };
      clearTimeout(debounceTimeout);
      debounceTimeout = setTimeout(later, wait);
    };
  }

  async function searchActors(query) {
    // Need at least 3 characters to search
    if (query.length < 3) {
      results.classList.add("hidden");
      results.innerHTML = "";
      return;
    }

    // Cancel previous request
    if (abortController) {
      abortController.abort();
    }
    abortController = new AbortController();

    try {
      const response = await fetch(
        `/api/search-actors?q=${encodeURIComponent(query)}`,
        {
          signal: abortController.signal,
        },
      );

      if (!response.ok) {
        results.classList.add("hidden");
        results.innerHTML = "";
        return;
      }

      const data = await response.json();

      if (!data.actors || data.actors.length === 0) {
        results.innerHTML =
          '<div class="px-4 py-3 text-sm text-gray-500">No accounts found</div>';
        results.classList.remove("hidden");
        return;
      }

      // Clear previous results
      results.innerHTML = "";

      // Create actor elements using DOM methods to prevent XSS
      data.actors.forEach((actor) => {
        const avatarUrl = actor.avatar || "/static/icon-placeholder.svg";
        const displayName = actor.displayName || actor.handle;

        // Create container div
        const resultDiv = document.createElement("div");
        resultDiv.className =
          "handle-result px-3 py-2 hover:bg-gray-100 cursor-pointer flex items-center gap-2";
        resultDiv.setAttribute("data-handle", actor.handle);

        // Create avatar image
        const img = document.createElement("img");
        // Validate URL scheme to prevent javascript: URLs
        if (
          avatarUrl &&
          (avatarUrl.startsWith("https://") || avatarUrl.startsWith("/static/"))
        ) {
          img.src = avatarUrl;
        } else {
          img.src = "/static/icon-placeholder.svg";
        }
        img.alt = ""; // Empty alt for decorative images
        img.width = 32;
        img.height = 32;
        img.className = "w-6 h-6 rounded-full object-cover flex-shrink-0";
        img.addEventListener("error", function () {
          this.src = "/static/icon-placeholder.svg";
        });

        // Create text container
        const textContainer = document.createElement("div");
        textContainer.className = "flex-1 min-w-0";

        // Create display name element
        const nameDiv = document.createElement("div");
        nameDiv.className = "font-medium text-sm text-gray-900 truncate";
        nameDiv.textContent = displayName; // textContent auto-escapes

        // Create handle element
        const handleDiv = document.createElement("div");
        handleDiv.className = "text-xs text-gray-500 truncate";
        handleDiv.textContent = "@" + actor.handle; // textContent auto-escapes

        // Assemble the elements
        textContainer.appendChild(nameDiv);
        textContainer.appendChild(handleDiv);
        resultDiv.appendChild(img);
        resultDiv.appendChild(textContainer);

        // Add click handler
        resultDiv.addEventListener("click", function () {
          input.value = actor.handle; // Use the actual handle from data, not DOM
          results.classList.add("hidden");
          results.innerHTML = "";
        });

        results.appendChild(resultDiv);
      });

      results.classList.remove("hidden");
    } catch (error) {
      if (error.name !== "AbortError") {
        console.error("Error searching actors:", error);
      }
    }
  }

  const debouncedSearch = debounce(searchActors, 300);

  input.addEventListener("input", function (e) {
    debouncedSearch(e.target.value);
  });

  // Hide results when clicking outside
  document.addEventListener("click", function (e) {
    if (!input.contains(e.target) && !results.contains(e.target)) {
      results.classList.add("hidden");
    }
  });

  // Show results again when input is focused
  input.addEventListener("focus", function () {
    if (results.innerHTML && input.value.length >= 3) {
      results.classList.remove("hidden");
    }
  });
})();
