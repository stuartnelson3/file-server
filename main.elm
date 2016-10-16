import Html exposing (..)
import Html.App as Html
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as Json exposing (..)
import Task


main =
  Html.program
    { init = init
    , view = view
    , update = update
    , subscriptions = subscriptions
    }

-- Model

type alias Model =
  { title : String
  , imgUrl : String
  , search : String
  , year : String
  , kind : String
  }


init : (Model, Cmd Msg)
init =
  (Model "" "" "" "" "", Cmd.none)


-- Update

type Msg
  = Search
  | NewInput String
  | FetchSucceed (List ApiResponse)
  | FetchFail Http.Error


update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    Search ->
      (model, searchApi model.search)

    NewInput search ->
      ({ model | search = search }, Cmd.none)

    FetchSucceed results ->
      let
          resp = (Maybe.withDefault (ApiResponse "" "" "" "" "") (List.head results))
      in
        ({ model | imgUrl = resp.poster, title = resp.title, year = resp.year, kind = resp.kind }, Cmd.none)

    FetchFail _ ->
      (model, Cmd.none)


-- View
view : Model -> Html Msg
view model =
  div []
    [ input [ type' "text", placeholder "Movie search term", onInput NewInput ] []
    , button [ onClick Search ] [ text "Search" ]
    , br [] []
    , img [ src model.imgUrl ] []
    , h1 [] [ text model.title ]
    , div []
      [ p [] [ text model.kind ]
      , p [] [ text model.year ]
      ]
    ]


-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
  Sub.none


-- HTTP

searchApi : String -> Cmd Msg
searchApi search =
  let
      url = "http://www.omdbapi.com/?s=" ++ search
  in
     Task.perform FetchFail FetchSucceed (Http.get decodeApiResponse url)

decodeApiResponse : Json.Decoder (List ApiResponse)
decodeApiResponse =
  -- How to get first index in search
  Json.at ["Search"] (Json.list responseDecoder)

type alias ApiResponse =
  { title : String
  , year : String
  , imdbID : String
  , kind : String
  , poster : String
  }

responseDecoder : Json.Decoder ApiResponse
responseDecoder =
  Json.object5 ApiResponse
    ("Title" := Json.string)
    ("Year" := Json.string)
    ("imdbID" := Json.string)
    ("Type" := Json.string)
    ("Poster" := Json.string)

