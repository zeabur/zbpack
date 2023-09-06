from django.shortcuts import render
from django.http.request import HttpRequest
from django.http.response import HttpResponse

from .forms import UploadForm
from .models import Profile


def profile(request: HttpRequest):
    if request.method == 'POST':
        form = UploadForm(request.POST, request.FILES)

        if form.is_valid():
            profile = Profile(image=request.FILES['image'])
            profile.save()
            return render(request, 'profile.html', {'profile': profile})
        else:
            return HttpResponse('Image upload failed')

    form = UploadForm()
    return render(request, 'profile.html', {'form': form})
