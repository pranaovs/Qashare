# DBMS Project: Finance Sharing Application

## Project Selection and Ideation

> Splitwise but FOSS
>
> Current splitwise application is riddled with ads and is not open source.

### Basic Idea

* Add accounts of users to share a particular set of expense
* Allow user accounts so that users can add expenses to their accounts
* Split an expense among users

### Feature set

1. No ads
2. FOSS with GPL v3 license
3. User accounts with email/password authentication
4. Addition of users to share expenses
5. Expenses can be added by other users with permission management.
Not only the owner of the expense can edit/delete it,
but others can also.
6. Expenses addition request can be sent to the owner of the group.
7. Uneven split of expenses among users
8. Final settlement of expenses
9. SQL database.
10. Ability to use local DB (SQLite) for shared expenses to keep application fully offline
11. Export data in various output formats (XML, JSON, CSV)
12. Optional: E2E encryption of expenses (with shared groups encryption)
13. Restrictive database access for end users

### Shared Expenses

* Create a group of users to share expenses
* New expenses are added to the group
* Uneven split of expenses among users
* Flags to denote that an entry hasn't been populated correctly/yet.
* Expense filtering
<!-- * Private expense (should be allowed) to  -->
* Permission management
  * Allow only owners to add, allow everyone to add, allow only owners to delete others expense, allow others to delete others expenses, edit permission, view permission, log permission, arbitrary expense addition/expense addition request.
  * Option to enable/disable permission management.
* Expense edit history required to make sure bad-actors do not edit an expense.
* Cloud based storage for sharing
* Final expense tally to show who owes how much.
  * Transaction logs for each user to show what a user has spent.

#### Workflow

1. Add a group with details
2. Invite users to the groups by selecting contacts/email/name
    * Online:
        1. Other users are added
    * Offline:
        1. Local SQLite databae/table is created
        2. Users can be added but locally
3. User 1 adds an expense:
    1. User selects contributor(s), contribution per user and which user(s) the expense has been done
        * Expense split can be set by amount/by percentage.
        * Autofilling the remaining amount possible
        * Equal split possible (quick toggle/button)
        * Flag an entry (lenders/borrowers) to note that the entry is not completed yet, and has to be filled in later.
    2. Data saved along with who added the expense
    3. Addition of images, location, description, date, time, type of expense (customizable)
        * Add option to set public/private mode. If an expense is private, the logs/transaction history is only visible to the lenders and borrowers.
4. View the total owed amount by each of them
5. Allow filtering expenses based on parties owed, parties lent, parties involved, date range, amount greater than, amount lesser than, amount in range.

### Private expenses for others

* Individual expenses can be added for other users
* It needs to be split unevenly among users
* Total can be tallied up on demand
* Markers are required to show which expenses are settled and which are not
* Logs need to be accessed to show who spend what
* Expenses need to be settled by others
* It is not shared

#### Workflow

1. User creates a spending entry
    * Provide amount spent, spent to which user(s) and lent amount per borrower
    * Add description, date, time, type of expense (customizable), location, images
2. Show amount lent on demand
    * Split the amount into amount lent per user
    * Show logs for each user spending

### Language choice

* Frontend (devices): Dart programming language with Flutter UI Framework
* Backend (to be run in servers and locally): Go

### Building workflow

#### Frontend

[ ] Create sample static data for user, groups, transaction list for a group

[ ] Visualize user info

[ ] Visualize group info

[ ] Visualize an expense

[ ] Create a new user

[ ] Create a new split group

[ ] Create a new expense

#### Backend

[ ] Send user info

[ ] Send group info

[ ] Send an expense info

[ ] Handle new group creation

[ ] Handle new expense addition

[ ] Handle new user creation

#### Database

[ ] Fetch expense details

[ ] Fetch group details

[ ] Fetch user details

[ ] New expense split creation

[ ] New group creation

[ ] New user creation

## Conversion of Project Idea to ER Diagram

<!-- TODO: add er diagram -->

## Conversion of ER diagram to UML Class Diagram

```mermaid
 erDiagram
     USERS o|--|| GROUPS : owns
     GROUPS }|--|| GROUP_MEMBERS: has
     GROUPS }o--|| EXPENSES : makes
     EXPENSES }|--|| EXPENSE_SPLITS: splits
     USERS {
         uuid user_id PK
         string name
         string email
         timestamp created_at
         string password_hash
     }
     GROUPS {
         uuid group_id PK
         string name
         string description
         uuid created_by FK "USERS"
         timestamp created_at
     }
     GROUP_MEMBERS {
         uuid user_id FK "USERS"
         uuid group_id FK "GROUPS"
         timestamp joined_at
     }
     EXPENSES {
         uuid group_id FK "GROUPS"
         uuid added_by FK "USERS"
         string title
         string description
         timestamp created_at
         float amount
         bool is_incomplete_amount
         bool is_incomplete_split
         float latitude
         float longitude
     }
     EXPENSE_SPLITS {
         uuid expense_id FK "EXPENSES"
         uuid user_id FK "USERS"
         float amount
         string role "paid/owes"
     }
```

## Conversion of UML Diagram to Database

Creation of database in MySQL:

```sql
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE groups (
    group_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_name TEXT NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users (user_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE group_members (
    user_id UUID REFERENCES users (user_id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups (group_id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE expenses (
    expense_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES groups (group_id) ON DELETE CASCADE,
    added_by UUID REFERENCES users (user_id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    is_incomplete_amount BOOLEAN DEFAULT FALSE,
    is_incomplete_split BOOLEAN DEFAULT FALSE,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION
);

CREATE TABLE expense_splits (
    expense_id UUID REFERENCES expenses (expense_id) ON DELETE CASCADE,
    user_id UUID REFERENCES users (user_id) ON DELETE CASCADE,
    amount DOUBLE PRECISION NOT NULL,
    user_role TEXT CHECK (user_role IN ('paid', 'owes')),
    PRIMARY KEY (expense_id, user_id)
);
```
