CREATE OR REPLACE FUNCTION public.safe_make_date(year bigint, month bigint, day bigint)
RETURNS date 
LANGUAGE plpgsql IMMUTABLE AS 
$$
DECLARE
	syear smallint;
	smonth smallint;
	sday smallint;
BEGIN
	BEGIN
		syear := year::smallint;
		smonth := month::smallint;
		sday := day::smallint;
    	RETURN make_date(syear, smonth, sday);
	EXCEPTION
	    WHEN others THEN 
    	    RETURN null;
	END;
END;
$$;