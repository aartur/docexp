## Users

To get the registered users, execute:

```python
from app import models

users = models.User.objects.all()
print(users[0])
```

### Database access

You can also inspect the database content directly:

```sql
SELECT * FROM user LIMIT 3
```