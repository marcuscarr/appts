from datetime import datetime, timedelta, timezone
from http import HTTPStatus
import requests

from init_db import initialize


def checkSuccess(r):
    print("Checking success")
    if r.status_code not in [HTTPStatus.CREATED, HTTPStatus.OK]:
        print(f"Unexpected response: Failed to create appointment: {r.text}")
    else:
        print(f"Created appointment: {r.json()}")


def checkFailure(r):
    print("Checking failure", r.status_code)
    if r.status_code not in [
        HTTPStatus.NOT_FOUND,
        HTTPStatus.BAD_REQUEST,
        HTTPStatus.INTERNAL_SERVER_ERROR,
    ]:
        print(f"Unexpected response: Created invalid appointment: {r.text}")
    else:
        print(f"Did not create invalid appointment: {r.text}")


def main():
    initialize()

    # Create a new appointment
    r = requests.post(
        "http://localhost:8080/appointments",
        json={
            "start_time": "2020-01-01T10:00:00-08:00",
            "end_time": "2020-01-01T10:30:00-08:00",
            "user_id": 1,
            "trainer_id": 2,
        },
    )

    checkSuccess(r)

    # ########################
    # # Appointment failures #
    # ########################

    # Attempt to create an appointment with a user that does not exist
    r = requests.post(
        "http://localhost:8080/appointments",
        json={
            "start_time": "2020-01-02T10:00:00-08:00",
            "end_time": "2020-01-02T10:30:00-08:00",
            "user_id": 100,
            "trainer_id": 1,
        },
    )
    checkFailure(r)

    # Attempt to create an appointment with a trainer that does not exist
    r = requests.post(
        "http://localhost:8080/appointments",
        json={
            "start_time": "2020-01-02T10:00:00-08:00",
            "end_time": "2020-01-02T10:30:00-08:00",
            "user_id": 1,
            "trainer_id": 100,
        },
    )
    checkFailure(r)

    # Attempt to create an appointment outside of business hours
    r = requests.post(
        "http://localhost:8080/appointments",
        json={
            "start_time": "2020-01-01T06:00:00-08:00",
            "end_time": "2020-01-01T06:30:00-08:00",
            "user_id": 1,
            "trainer_id": 2,
        },
    )

    checkFailure(r)

    # Attempt to create an appointment with a start time after the end time
    r = requests.post(
        "http://localhost:8080/appointments",
        json={
            "start_time": "2020-01-01T11:30:00-08:00",
            "end_time": "2020-01-01T11:00:00-08:00",
            "user_id": 1,
            "trainer_id": 2,
        },
    )

    ############################
    # Appointment availability #
    ############################

    # Get a trainer
    r = requests.get("http://localhost:8080/trainers/1")
    print(f"Trainer: {r.json()}")

    # # Get a trainer's appointments
    r = requests.get("http://localhost:8080/trainers/1/appointments")
    appointments = r.json()
    apptSet = {a["start_time"] for a in appointments}

    # Get the availability of a trainer
    r = requests.get(
        "http://localhost:8080/trainers/1/appointments/available",
        params={
            "starts_at": "2019-01-24",
            "ends_at": "2019-01-26",
        },
    )
    available = r.json()
    availableSet = set(available["available"])

    isoFormat = "%Y-%m-%dT%H:%M:%S%z"
    first = datetime.strptime(appointments[0]["start_time"], isoFormat)
    last = datetime.strptime(appointments[-1]["start_time"], isoFormat)

    tz = timezone(timedelta(hours=-8))
    firstFormatted = first.astimezone(tz)
    lastFormatted = last.astimezone(tz)

    print(f"{len(apptSet)} appointments; {len(availableSet)} available")
    print(f"Appointments & available intersection: {apptSet & availableSet}")
    print(f"First available: {available['available'][0]}")
    print(f"First booked: {firstFormatted}")
    print(f"Last available: {available['available'][-1]}")
    print(f"Last booked: {lastFormatted}")


if __name__ == "__main__":
    main()
