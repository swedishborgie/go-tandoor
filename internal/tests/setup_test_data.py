from cookbook.models import Space, UserSpace, Household, InventoryLocation
from django_scopes import scopes_disabled
from django.contrib.auth.models import Group
from django.contrib.auth import get_user_model

User = get_user_model()
u = User.objects.get(username='testadmin')

with scopes_disabled():
    s, _ = Space.objects.get_or_create(name='TestSpace', defaults={'created_by': u})
    s.created_by = u
    s.save()
    h, _ = Household.objects.get_or_create(name='Test Household', space=s)
    us, _ = UserSpace.objects.get_or_create(user=u, space=s, defaults={'active': True})
    us.active = True
    us.household = h
    us.save()
    il, _ = InventoryLocation.objects.get_or_create(
        name='Test Location',
        household=h,
        defaults={'is_freezer': False, 'created_by': u, 'space': s}
    )
    g_user = Group.objects.get(name='user')
    us.groups.add(g_user)
    g_admin = Group.objects.get(name='admin')
    us.groups.add(g_admin)
    print('setup done')
