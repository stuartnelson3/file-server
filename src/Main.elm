-- External Imports
import Navigation

-- Internal Imports
import Parsing
import Views
import Api
import Types exposing (..)

main =
  Navigation.program Parsing.urlParser
    { init = init
    , view = Views.view
    , update = update
    , urlUpdate = urlUpdate
    , subscriptions = subscriptions
    }


init : Route -> (Model, Cmd Msg)
init route =
  urlUpdate route (Model [] (ApiResponse "" "" (MovieResponse "" "" "" "" "")) route)

-- Update


update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    MoviesFetchSucceed movies ->
        ({ model | movies = movies }, Cmd.none)

    MovieFetchSucceed movie ->
        ({ model | movie = movie }, Cmd.none)

    FetchFail _ ->
      ({model | route = NotFound }, Cmd.none)


urlUpdate : Route -> Model -> (Model, Cmd Msg)
urlUpdate route model =
  let
      cmd =
        case route of
          Movie imdbID ->
            let
              one = Debug.log "urlUpdate: imdbID" imdbID
            in
              Api.singleMovieSearch imdbID

          Movies ->
            Api.searchApi

          TopLevel ->
            Navigation.modifyUrl "/#/movies"

          _ ->
            Cmd.none
  in
    ({model | route = route }, cmd)


-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
  Sub.none


