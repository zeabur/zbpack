# Static and Media Files in Django

Example repo for the [Working with Static and Media Files in Django](https://testdriven.io/blog/django-static-files/) article.

## Getting Started

1. Clone down the repo
1. Create and activate a virtual environment
1. Install the dependencies:

    ```sh
    $ pip install -r requirements.txt
    ```

1. Apply the migrations:

    ```sh
    $ python manage.py migrate
    ```

### Development Example

```sh
$ python manage.py runserver
```

### Production Example

1. Set `DEBUG` to `False` in the *settings.py* file
1. Then, collect the static files and run Gunicorn:

    ```sh
    $ python manage.py collectstatic
    $ gunicorn core.wsgi:application -w 1
    ```

This *DOES NOT* use WhiteNoise to serve up the static files.
