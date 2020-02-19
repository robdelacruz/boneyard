import datetime

def isodt_to_date(isodt):
    try:
        date = datetime.datetime.strptime(isodt, "%Y-%m-%d")
    except:
        date = None
    return date


def date_to_isodt(date):
    return date.strftime("%Y-%m-%d")


def dateparts_to_isodt(year, month, day):
    date = datetime.date(year, month, day)
    return date_to_isodt(date)


def today_isodt():
    return datetime.datetime.now().strftime("%Y-%m-%d")


def isodt_to_longfmt(isodt):
    """ Return long format: Ex. Mon 01 Apr 2019 """
    date = isodt_to_date(isodt)
    return f"{date:%a %d %b %Y}"

def isodt_to_shortfmt(isodt):
    """ Return long format: Ex. Mon 01 Apr """
    date = isodt_to_date(isodt)
    return f"{date:%a %d %b}"
