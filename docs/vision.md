# Vision: Home Cooking Ecosystem

A personal ecosystem of apps for managing everything related to home cooking —
recipes, shopping, pantry inventory, and meal planning — with purpose-built
interfaces for each context.

## Apps

### Shopping List (Mobile)
Android app for managing shopping lists on the go. Offline-first with sync.
Migrated from the standalone shoppingList project.

### Cooking Assistant (Mobile)
Android app for recipe management and step-by-step cooking guidance.
Migrated from the standalone cookingApp project.

### Pantry Inventory (Tablet)
Tablet-optimized interface for tracking what's in the pantry. Scan or enter
items, track quantities, get notified when things are running low. Connects
to recipes (what can I make with what I have?) and shopping lists (what do I
need to buy?).

### Shopping Planner (Desktop)
Desktop interface for shopping list management with keyboard/mouse efficiency.
Link deals, plan shopping trips, manage lists from a computer. More ergonomic
than touch for bulk planning tasks.

### Emacs Integration
Elisp package and/or org-mode export that lets you author shopping lists in
org-mode and push them to the API. For when you're already in emacs and don't
want to context switch.

### Recipe Database
Recipe storage and retrieval, potentially as its own frontend or integrated
into the cooking assistant. The API serves recipes to any client that needs
them.

## How They Connect

```
                    ┌─────────────┐
                    │   Go API    │
                    │  (single    │
                    │   server)   │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
         ┌────┴────┐  ┌───┴───┐  ┌────┴────┐
         │ Recipes │  │ Lists │  │ Pantry  │
         └────┬────┘  └───┬───┘  └────┬────┘
              │            │            │
              └────────────┼────────────┘
                           │
        ┌──────────┬───────┼───────┬──────────┐
        │          │       │       │          │
    ┌───┴───┐ ┌───┴──┐ ┌──┴──┐ ┌─┴────┐ ┌───┴───┐
    │Cooking│ │ Shop │ │Tablet│ │ Desk │ │ Emacs │
    │  App  │ │ App  │ │ App │ │  App │ │       │
    └───────┘ └──────┘ └─────┘ └──────┘ └───────┘
```

- Recipes feed shopping lists (generate a list from a meal plan)
- Pantry inventory subtracts what you have from what you need
- All clients see the same data through the shared API
- Each client is optimized for its form factor and use case

## Design Principles

- **Offline-first**: Mobile clients work without connectivity and sync when available
- **Single API**: One Go server process handles all domains, deployed as one container
- **Protobuf contracts**: Type definitions live in `.proto` files, generated for each language
- **Purpose-built interfaces**: Each app is optimized for its context rather than one app trying to do everything
- **Personal scale**: This is a single-user ecosystem, not a SaaS product. Simplicity over scalability.
