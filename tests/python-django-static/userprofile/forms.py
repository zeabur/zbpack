from django import forms


class UploadForm(forms.Form):
    image = forms.FileField(label='Upload a picture')
