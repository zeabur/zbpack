from django.urls import path

from .views import profile


urlpatterns = [
    path('profile/', profile, name='profile'),
]
